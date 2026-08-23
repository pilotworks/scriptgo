// @expect: Bob
// @expect: default_country
// @expect: default_theme
type Settings = {
    theme?: string;
};

type Profile = {
    username: string;
    country?: string;
    settings?: Settings;
};

type AppData = {
    user: {
        profile: Profile;
    };
};

const obj: AppData = {
    user: {
        profile: {
            username: "Bob",
        },
    },
};

const { user: { profile: { username, country = "default_country", settings: { theme = "default_theme" } = { theme: "default_theme" } } } } = obj;
console.log(username);
console.log(country);
console.log(theme);
