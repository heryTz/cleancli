# Clean CLI

It is a simple cache cleaner command line. It supports Mac, Linux and Windows.

![](https://github.com/heryTz/cleancli/blob/main/demo.gif)

## Installation

From binary

Download the archive matching your OS and architecture from the [latest release](https://github.com/heryTz/cleancli/releases/latest), then:

```bash
tar -xzf cleancli_<version>_<os>_<arch>.tar.gz
chmod +x ./cleancli
./cleancli
```

From source

```bash
git clone https://github.com/heryTz/cleancli
cd cleancli
go build .
chmod +x ./cleancli
./cleancli
```

## Features

- [x] Scan and list cache entries with their size
- [x] Select one, multiple, or all entries to delete
- [x] Support Mac, Linux and Windows
- [ ] Show absolute path of each cache item
