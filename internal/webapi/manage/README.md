# logtail manage page front end project

## Usage

### install

```bash
yarn install
```

### dev

```bash
yarn run logtail:run # start logtail local server as backend
yarn run dev # start front end dev server
```

Front end dev server will running at [http://localhost:3000](http://localhost:3000) by default.

### build

```bash
yarn run logtail:sync
```

Above command will build project and then copy output to target. After sync done, you need restart backend server to make it effective. (stop and rerun `yarn run logtail:run`)
