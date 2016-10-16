# Sandbox server for Torigoya
Server backend for Torigoya  
Sandbox environment is developed on [yutopp/awaho](https://github.com/yutopp/awaho)

# Requrement
- golang >= 1.7.1
- Awaho

# Development
See [wiki](https://github.com/yutopp/torigoya_cage/wiki)(Japanese)

# Setup
First, you must create `config.yml` on `app` directory.  
See `config.yml.template`

## Ubuntu(14.04)
### Files(optional)
If you want to use torigoya deb repository, run `sudo cp ./files/torigoya-packages.list /etc/apt/sources.list.d/torigoya-packages.list`.

### Build
```
./build.sh
```
Then, run `./bin/cage.server` to host Cage.

## License
Boost License Version 1.0
