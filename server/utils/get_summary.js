// 私以外私じゃないの
const https = require('https')

const data = JSON.stringify({
  todo: 'Buy the milk'
})

const options = {
  hostname: "api-3moji.herokuapp.com",
  port: 443, method: 'POST',
  path: "/api/v1/summary/",
  headers: {
    'Content-Type': 'application/json',
    'Content-Length': data.length
  },
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

req.write(data)
req.end()
