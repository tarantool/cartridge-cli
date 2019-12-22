local checks = require('checks')
local errors = require('errors')
local decnumber = require('ldecnumber')
local cartridge = require('cartridge')

local err_max_balance = errors.new_class("Max balance error")


local function update_balance(balance, amount)
    -- Converts string to decimal object.
    local balance_decimal = decnumber.tonumber(balance)

    local maximum = cartridge.config_get_readonly('max-balance')
    if maximum and tonumber(amount) > tonumber(maximum) then
        return nil, err_max_balance:new("Maximum is "..tostring(maximum))
    end

    balance_decimal = balance_decimal + amount
    if balance_decimal:isnan() then
        error('Invalid amount')
    end

    -- Rounds up to 2 decimal places and converts back to string.
    return balance_decimal:rescale(-2):tostring()
end

local function init_spaces()
    local customer = box.schema.space.create(
            'customer', {
            format = {
                {'customer_id', 'unsigned'},
                {'bucket_id', 'unsigned'},
                {'name', 'string'},
            },
            if_not_exists = true,
        })
        customer:create_index('customer_id', {
            parts = {'customer_id'},
            if_not_exists = true,
        })
        customer:create_index('bucket_id', {
            parts = {'bucket_id'},
            unique = false,
            if_not_exists = true,
        })

        local account = box.schema.space.create('account', {
            format = {
                {'account_id', 'unsigned'},
                {'customer_id', 'unsigned'},
                {'bucket_id', 'unsigned'},
                {'balance', 'string'},
                {'name', 'string'},
            },
            if_not_exists = true,
        })
        account:create_index('account_id', {
            parts = {'account_id'},
            if_not_exists = true,
        })
        account:create_index('customer_id', {
            parts = {'customer_id'},
            unique = false,
            if_not_exists = true,
        })
        account:create_index('bucket_id', {
            parts = {'bucket_id'},
            unique = false,
            if_not_exists = true,
        })
end

local function customer_add(customer)
    customer.accounts = customer.accounts or {}

    box.begin()
    box.space.customer:insert({
        customer.customer_id,
        customer.bucket_id,
        customer.name
    })
    for _, account in pairs(customer.accounts) do
        box.space.account:insert({
            account.account_id,
            customer.customer_id,
            customer.bucket_id,
            '0.00',
            account.name or "default_name"..tostring(account.account_id)
        })
    end
    box.commit()
    return true
end

--- Adds amount to the customer's balance.
-- Rounds up balance to 2 decimal places.
-- Amount can be negative.
local function customer_update_balance(customer_id, account_id, amount)
    checks('number', 'number', 'string')

    local account = box.space.account:get(account_id)
    if account == nil then
        return nil
    end

    -- Checks account's validity.
    if account.customer_id ~= customer_id then
        error('Invalid account_id')
    end

    local new_balance, err = update_balance(account.balance, amount)

    if err ~= nil then
        return nil, err
    end

    box.space.account:update({ account_id }, {
        { '=', 4, new_balance }
    })

    return new_balance
end

local function customer_lookup(customer_id)
    checks('number')

    local customer = box.space.customer:get(customer_id)
    if customer == nil then
        return nil
    end
    customer = {
        customer_id = customer.customer_id;
        name = customer.name;
    }
    local accounts = {}
    for _, account in box.space.account.index.customer_id:pairs(customer_id) do
        table.insert(accounts, {
            account_id = account.account_id;
            name = account.name;
            balance = account.balance;
        })
    end
    customer.accounts = accounts;
    return customer
end

local exported_functions = {
    customer_add = customer_add,
    customer_lookup = customer_lookup,
    customer_update_balance = customer_update_balance,
}

local function init(opts)
    if opts.is_master then
        init_spaces()

        for name in pairs(exported_functions) do
            box.schema.func.create(name, {if_not_exists = true})
            box.schema.role.grant('public', 'execute', 'function', name, {if_not_exists = true})
        end
    end

    for name, func in pairs(exported_functions) do
        rawset(_G, name, func)
    end

    return true
end


return {
    role_name = 'storage',
    init = init,
    utils = {
        update_balance = update_balance
    },
    dependencies = {
        'cartridge.roles.vshard-storage',
    },
}
