const fs = require('fs');
const { resolve } = require('path');

function copy(src, dist) {
  fs.createReadStream(src).pipe(fs.createWriteStream(dist));
}

copy(resolve(__dirname, '../dist/index.html'), resolve(__dirname, '../../manage.html'))