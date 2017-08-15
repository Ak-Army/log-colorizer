# log-colorizer
A simple devops utility to emphasize / colorize certain parts of a log file. Used with grep and tail unix utilities.

# Installation
```
go get github.com/Ak-Army/log-colorizer
```
Customize `.config.json` and put it in the directory as `.config.json` where log-colorizer binary is found
Alternatively you can provide a path to the config file using `-c` flag.

# Usage
```
tail -f log_file.log | log-colorizer
```

