const fs = require("fs");

const responses = JSON.parse(fs.readFileSync("out.json"));
const times = responses.times;

for (let k in times) {
  const at = times[k];
  console.log(k, at);
}
