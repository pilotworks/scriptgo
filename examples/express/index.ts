import { Handler } from "./types";
import { Application } from "./application";
import { Router as RouterClass } from "./router";
import { json as expressJson, urlencoded as expressUrlencoded } from "./middleware";

export * from "./types";
export * from "./request";
export * from "./response";
export * from "./router";
export * from "./middleware";
export * from "./application";

export function express(): Application {
    return new Application();
}

export namespace express {
    export function json(): Handler {
        return expressJson();
    }
    export function urlencoded(): Handler {
        return expressUrlencoded();
    }
    export function Router(): RouterClass {
        return new RouterClass();
    }
}

export default express;
