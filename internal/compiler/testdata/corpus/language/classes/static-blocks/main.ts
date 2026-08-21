class ServerConfig {
  static port: number = 3000;
  static host: string = "localhost";
  static url: string = "";
  static isProduction: boolean = false;

  static {
    ServerConfig.port = 8080;
    ServerConfig.url = "http://" + ServerConfig.host + ":" + ServerConfig.port;
  }

  static counter: number = 10;
  static {
    this.counter = this.counter * 2;
    this.isProduction = true;
  }
}

console.log(ServerConfig.port);
console.log(ServerConfig.host);
console.log(ServerConfig.url);
console.log(ServerConfig.counter);
console.log(ServerConfig.isProduction);
