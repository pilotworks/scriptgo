// @expect: Action: CREATE
// @expect: Updating record...
// @expect: Deleting record...
// @expect: Success!
enum Command {
    Create = "CREATE",
    Update = "UPDATE",
    Delete = "DELETE"
}

function processCommand(cmd: Command): string {
    switch (cmd) {
        case Command.Create:
            return "Action: " + cmd;
        case Command.Update:
            return "Updating record...";
        case Command.Delete:
            return "Deleting record...";
    }
}

console.log(processCommand(Command.Create));
console.log(processCommand(Command.Update));
console.log(processCommand(Command.Delete));

const current: Command = Command.Create;
if (current === Command.Create) {
    console.log("Success!");
}
