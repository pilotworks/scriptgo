class BaseConfig {
  static defaultPort: number = 8080;
  static defaultHost: string = "localhost";

  static getUrl(): string {
    return "http://" + BaseConfig.defaultHost + ":" + BaseConfig.defaultPort;
  }
}

class AppConfig extends BaseConfig {
  static apiVersion: string = "v1";

  static getApiEndpoint(path: string): string {
    return BaseConfig.getUrl() + "/" + AppConfig.apiVersion + "/" + path;
  }
}

console.log(BaseConfig.defaultPort);
console.log(BaseConfig.defaultHost);
console.log(BaseConfig.getUrl());

console.log(AppConfig.apiVersion);
console.log(AppConfig.getApiEndpoint("users"));
console.log(AppConfig.getApiEndpoint("orders"));
