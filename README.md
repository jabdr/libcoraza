# C library for OWASP Coraza Web Application Firewall v3

Welcome to libcoraza, the C library for OWASP Coraza Web Application Firewall. Because [Coraza](https://github.com/corazawaf/coraza) has made in golang, if you want to embed in any kind of C application, you will need this library.

## Prerequisites

* a C compiler:
  * gcc or
  * clang
  * mingw (on windows)
* Golang compiler v1.21+
* cmake 3.24+
* make

## Download

Download the library source:

```
git clone https://github.com/corazawaf/libcoraza libcoraza
```

## Build source on Linux & Mac

```
cd libcoraza
mkdir build
cd build
cmake -DCMAKE_EXPORT_COMPILE_COMMANDS:BOOL=TRUE -B build -G "Unix Makefiles" ..
make all
sudo make install
```

## Build source on Windows

```
cd libcoraza
mkdir build
cd build
cmake -DCMAKE_EXPORT_COMPILE_COMMANDS:BOOL=TRUE -B build -G "MinGW Makefiles" .
make all
sudo make install
```

## Run test

If you want to try the given example, try:

```
#after make all inside build/
make test
```

## Others

If you didn't installed the builded library (skipped the `sudo make install` step), you should set the LD_LIBRARY_PATH:

```
export LD_LIBRARY_PATH=../:$LID_LIBRARY_PATH
```
