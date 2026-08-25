// @expect: #ffffff
// @expect: #000000
// @expect: #888888
type Theme = "light" | "dark" | "system";

function getBackground(theme: Theme): string {
    switch (theme) {
        case "light":
            return "#ffffff";
        case "dark":
            return "#000000";
        case "system":
            return "#888888";
    }
}

console.log(getBackground("light"));
console.log(getBackground("dark"));
console.log(getBackground("system"));
