// @expect: 8080
// @expect: localhost
// @expect: true
const config = {
    server_port: 8080,
    server_host: "localhost",
    is_active: true
};

const { server_port: port, server_host: host, is_active: active } = config;

console.log(port);
console.log(host);
console.log(active);
