# Sandbox environment for Torigoya
**!!under construction!!**
Currently, this program may harm your computer. PLEASE EXECUTE THIS PROGRAM IN THE VIRTUAL ENVIRONMENT.

## Development
git clone git@github.com:yutopp/torigoya_proc_profiles.git
git clone git@github.com:yutopp/torigoya_package_scripts.git
git clone git@github.com:yutopp/torigoya_factory.git

You can use Vagrantfile for debbuging.
exec `vagrant up` and `vagrant ssh`, then `cd /vagrant` and please use `host.*.sh` scripts.

## testing
docker.run_core_test.sh

docker.run_system_test.sh
docker.run_system_test.sh remote

## Requirement
[Docker](http://www.docker.com/ "Docker")
[Vagrant](http://www.vagrantup.com/ "Vagrant")(recommended)

## License

Boost License Version 1.0