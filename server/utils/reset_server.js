// 私以外私じゃないの
const https = require('https')


/*
const options = {
  hostname: "api-3moji.herokuapp.com",
  port: 443, method: 'POST',
  path: "/debug/reset_redis",
}

const req = https.request(options, res => {
  if (res.statusCode !== 200) return console.log(`statusCode: ${res.statusCode}`);

  res.on('data', d => {
    process.stdout.write(d)
  })
})

req.on('error', error => {
  console.error(error)
})

req.end()
*/
