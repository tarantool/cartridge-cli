cmake_minimum_required(VERSION 2.8 FATAL_ERROR)

project(cartridge-cli C)
if(NOT CMAKE_BUILD_TYPE)
    set(CMAKE_BUILD_TYPE Debug)
endif()

set(CMAKE_MODULE_PATH "${CMAKE_CURRENT_SOURCE_DIR}/cmake" ${CMAKE_MODULE_PATH})

# Find Tarantool and Lua dependecies
set(TARANTOOL_FIND_REQUIRED ON)
find_package(Tarantool)

enable_testing()

set (LUA_PATH "LUA_PATH=${PROJECT_SOURCE_DIR}/?.lua\\;${PROJECT_SOURCE_DIR}/?/init.lua\\;\\;")
set (LUA_SOURCE_DIR "LUA_SOURCE_DIR=${PROJECT_SOURCE_DIR}")

## Install ####################################################################
###############################################################################

install(
  CODE "execute_process(
    COMMAND ${CMAKE_CURRENT_SOURCE_DIR}/generate_binary.py
            --source-dir ${CMAKE_CURRENT_SOURCE_DIR}
            --dest-file ${TARANTOOL_INSTALL_BINDIR}/cartridge
            --module-name cartridge-cli
  )"
)
