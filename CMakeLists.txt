cmake_minimum_required(VERSION 3.24)

find_package(Git REQUIRED)

execute_process(
  COMMAND ${GIT_EXECUTABLE} describe --tags --dirty
  WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
  OUTPUT_VARIABLE PROJECT_VERSION
  RESULT_VARIABLE PROJECT_VERSION_ERROR_CODE
  OUTPUT_STRIP_TRAILING_WHITESPACE
)

if(NOT ${PROJECT_VERSION_ERROR_CODE} EQUAL 0)
  set(PROJECT_VERSION 0.0.0)
endif()

project(libcoraza VERSION ${PROJECT_VERSION})

include(GNUInstallDirs)
include(CTest)

find_program(GO_BIN NAMES go REQUIRED)
find_package(Doxygen REQUIRED)

file(GLOB_RECURSE GO_SRCS ${CMAKE_CURRENT_SOURCE_DIR}/libcoraza/*.go)
set(GO_LIB ${CMAKE_STATIC_LIBRARY_PREFIX}coraza${CMAKE_STATIC_LIBRARY_SUFFIX})
set(GO_HEADER ${CMAKE_SHARED_LIBRARY_PREFIX}coraza.h)
set(OUTPUT_HEADER coraza.h)

add_custom_command(OUTPUT ${CMAKE_CURRENT_BINARY_DIR}/${GO_LIB} ${CMAKE_CURRENT_BINARY_DIR}/${GO_HEADER}
  DEPENDS ${GO_SRCS} ${CMAKE_CURRENT_SOURCE_DIR}/go.mod ${CMAKE_CURRENT_SOURCE_DIR}/go.sum
  WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
  COMMAND ${GO_BIN} build -buildmode=c-archive
  -o "${CMAKE_CURRENT_BINARY_DIR}/${GO_LIB}"
  ${CMAKE_GO_FLAGS} ${CMAKE_CURRENT_SOURCE_DIR}/libcoraza/...
  COMMENT "Building Go static library")

add_custom_target(libcoraza_static DEPENDS ${CMAKE_CURRENT_BINARY_DIR}/${GO_LIB} ${CMAKE_CURRENT_BINARY_DIR}/${GO_HEADER})

add_custom_command(OUTPUT ${CMAKE_CURRENT_BINARY_DIR}/coraza/${OUTPUT_HEADER}
  DEPENDS libcoraza_static ${CMAKE_CURRENT_BINARY_DIR}/${GO_HEADER}
  COMMAND ${CMAKE_COMMAND} -E copy ${CMAKE_CURRENT_BINARY_DIR}/${GO_HEADER} ${CMAKE_CURRENT_BINARY_DIR}/coraza/${OUTPUT_HEADER}
  COMMENT "Copy header to destination")

add_custom_target(coraza_header DEPENDS ${CMAKE_CURRENT_BINARY_DIR}/coraza/${OUTPUT_HEADER})

add_library(coraza_static STATIC IMPORTED GLOBAL)
add_dependencies(coraza_static libcoraza_static coraza_header)
set_target_properties(coraza_static
  PROPERTIES
  IMPORTED_LOCATION ${CMAKE_CURRENT_BINARY_DIR}/${GO_LIB}
  INTERFACE_INCLUDE_DIRECTORIES ${CMAKE_CURRENT_BINARY_DIR})

add_library(coraza SHARED ${CMAKE_CURRENT_BINARY_DIR}/${GO_LIB})
target_link_libraries(coraza PRIVATE "$<LINK_LIBRARY:WHOLE_ARCHIVE,coraza_static>")
add_dependencies(coraza coraza_static)
set_target_properties(coraza
  PROPERTIES
  LINKER_LANGUAGE CXX
  INTERFACE_INCLUDE_DIRECTORIES ${CMAKE_CURRENT_BINARY_DIR})

set(DOXYGEN_EXCLUDE_SYMBOLS _GoString_, GoInterface, GoSlice)
set(DOXYGEN_GENERATE_HTML YES)
set(DOXYGEN_GENERATE_MAN YES)
set(DOXYGEN_FILE_PATTERNS .h)
set(DOXYGEN_TOC_INCLUDE_HEADINGS 5)
doxygen_add_docs(docs ${CMAKE_CURRENT_BINARY_DIR} ALL)
set_target_properties(docs PROPERTIES DEPENDS coraza_header)

add_executable(simple_get_shared tests/simple_get.c)
target_link_libraries(simple_get_shared coraza)

add_executable(simple_get_static tests/simple_get.c)
target_link_libraries(simple_get_static coraza_static)

add_test(NAME simple_get_static COMMAND simple_get_static)
set(TEST simple_get_static PROPERTY PASS_REGULAR_EXPRESSION "Transaction disrupted with status 403")

add_test(NAME simple_get_shared COMMAND simple_get_shared)
set(TEST simple_get_shared PROPERTY PASS_REGULAR_EXPRESSION "Transaction disrupted with status 403")

add_test(NAME go
         COMMAND ${GO_BIN} test ${CMAKE_CURRENT_SOURCE_DIR}/libcoraza/...)
set(TEST go PROPERTY FAIL_REGULAR_EXPRESSION  "^FAIL")

install(TARGETS coraza)
install(DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}/coraza/ DESTINATION ${CMAKE_INSTALL_INCLUDEDIR}/coraza)
install(FILES ${CMAKE_CURRENT_BINARY_DIR}/${GO_LIB} TYPE LIB)
install(DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}/html/ TYPE DOC)
install(DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}/man/ TYPE MAN)
