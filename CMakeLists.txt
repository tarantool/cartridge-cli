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

# Create tempdir
execute_process(
  COMMAND mktemp -d
  OUTPUT_STRIP_TRAILING_WHITESPACE
  OUTPUT_VARIABLE TMPDIR
)

# Generate cartridge executable in tmpdir
execute_process(
  COMMAND ${TARANTOOL} ${CMAKE_CURRENT_SOURCE_DIR}/compile_executable.lua
          --source ${CMAKE_CURRENT_SOURCE_DIR}
          --dest ${TMPDIR}/cartridge
          --module cartridge-cli
  RESULT_VARIABLE ret
)
if(ret EQUAL "1")
    message( FATAL_ERROR "Failed to compile cartridge executable")
endif()

install(
  PROGRAMS ${TMPDIR}/cartridge
  DESTINATION ${TARANTOOL_INSTALL_BINDIR}
)
