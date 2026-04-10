# fireball - the investment monitoring app
## About
Fireball is an investment app for investors that allows them to plan and track their investments, control dividends, and help them make investment decisions. Works with **T-Invest-Api**.
## Installation guide
Web app is not yet available in net, currently working on it...

To try app locally clone repo to your machine. App was developed on Ubuntu and MacOS, so Windows is not supported for dev environment. Make sure you you have [go](https://go.dev/doc/install) and [node](https://nodejs.org/en/download) installed on your machine. Setup scripts are written in makefile, so if you for some reason don't have **make** utility, get it installed too
Also, we use image of [Redis](hub.docker.com/layers/library/redis/7.0-alpine/images/sha256-58ff33524a6664b35e50c1a243f07096eb05b479dc65e06f3592197353ca10fa), so [docker] is necessary for correct work

* After cloning repo, create the **.env** file in the root of the repo. Contents example:
```
T_INVEST_URL="https://invest-public-api.tbank.ru"
REDIS_URL="your-redis-url"
REDIS_TTL="24"
SESSION_SECRET="1234567890qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM64"
PORT="8080"
READ_TIMEOUT="10"
WRITE_TIMEOUT="10"
IDLE_TIMEOUT="30"
```
* Don't forget to paste correct redis-url and generate and paste 32 bytes secret. You can use `openssl rand -hex 64`
* Download dependencies using command `make setup-environment` from the root of the repo
* Type and enter `make build-frontend` to build required assets for frontend
    * For deploying web resources **Vite** is used. It runs on 20s version of Node
* Enter `make start-deps` for building required docker containers
* Run `make run-server` to wind up the backend. App will be served on `http://localhost:8080/`
* If you need a development media for real-time edit of frontend code, use `make dev`. Vite will serve app on `http://localhost:3000/`. Don't forget run the server in the previous step, if you want not only frontend page, but fully functional app 

Contact us, authors, for possible suggestions or complaints: 
* [CatSprite](https://github.com/CatSprite-dev)
* [ar3ty](https://github.com/ar3ty)