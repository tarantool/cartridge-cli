cmake_minimum_required(VERSION 2.8 FATAL_ERROR)

project(cartridge-cli C)
if(NOT CMAKE_BUILD_TYPE)
    set(CMAKE_BUILD_TYPE Debug)
endif()

set(CMAKE_MODULE_PATH "${CMAKE_CURRENT_SOURCE_DIR}/cmake" ${CMAKE_MODULE_PATH})

if (NOT TARANTOOL_INSTALL_LUADIR)
    set(TARANTOOL_INSTALL_LUADIR "${CMAKE_INSTALL_PREFIX}/share/tarantool")
endif()

if (NOT TARANTOOL_INSTALL_BINDIR)
    set(TARANTOOL_INSTALL_BINDIR "${CMAKE_INSTALL_PREFIX}/bin")
endif()

enable_testing()

set (LUA_PATH "LUA_PATH=${PROJECT_SOURCE_DIR}/?.lua\\;${PROJECT_SOURCE_DIR}/?/init.lua\\;\\;")
set (LUA_SOURCE_DIR "LUA_SOURCE_DIR=${PROJECT_SOURCE_DIR}")

## VERSION ####################################################################
###############################################################################

if (NOT VERSION)
  execute_process(
    COMMAND git describe --tags --always
    WORKING_DIRECTORY ${PROJECT_SOURCE_DIR}
    OUTPUT_STRIP_TRAILING_WHITESPACE
    OUTPUT_VARIABLE VERSION
    ERROR_QUIET
  )
endif()

if (NOT VERSION)
  message(FATAL_ERROR "VERSION is not provided")
endif()

configure_file (
  "${PROJECT_SOURCE_DIR}/${PROJECT_NAME}/VERSION.lua.in"
  "${PROJECT_SOURCE_DIR}/${PROJECT_NAME}/VERSION.lua"
)

## Install ####################################################################
###############################################################################

install(
  FILES ${CMAKE_CURRENT_SOURCE_DIR}/cartridge-cli.lua
  DESTINATION ${TARANTOOL_INSTALL_LUADIR}/
)

install(
  DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}/templates
            ${CMAKE_CURRENT_SOURCE_DIR}/${PROJECT_NAME}
  USE_SOURCE_PERMISSIONS
  DESTINATION ${TARANTOOL_INSTALL_LUADIR}/
  PATTERN "*.in" EXCLUDE
)

install(
  PROGRAMS ${CMAKE_CURRENT_SOURCE_DIR}/cartridge
  DESTINATION ${TARANTOOL_INSTALL_BINDIR}
)
