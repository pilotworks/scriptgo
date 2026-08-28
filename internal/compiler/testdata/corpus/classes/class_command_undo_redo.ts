// @expect: Hello World!
// @expect: Hello World
// @expect: Hello 
// @expect: Hello World
// @expect: Hello World TypeScript
interface Command {
    execute(): void;
    undo(): void;
}

class TextEditor {
    private content: string = "";

    append(text: string): void {
        this.content += text;
    }

    delete(count: number): void {
        if (count > this.content.length) count = this.content.length;
        this.content = this.content.substring(0, this.content.length - count);
    }

    getText(): string {
        return this.content;
    }
}

class AppendCommand implements Command {
    private editor: TextEditor;
    private text: string;

    constructor(editor: TextEditor, text: string) {
        this.editor = editor;
        this.text = text;
    }

    execute(): void {
        this.editor.append(this.text);
    }

    undo(): void {
        this.editor.delete(this.text.length);
    }
}

class CommandManager {
    private history: Command[] = [];
    private redoStack: Command[] = [];

    execute(cmd: Command): void {
        cmd.execute();
        this.history.push(cmd);
        this.redoStack = [];
    }

    undo(): void {
        const cmd = this.history.pop();
        if (cmd) {
            cmd.undo();
            this.redoStack.push(cmd);
        }
    }

    redo(): void {
        const cmd = this.redoStack.pop();
        if (cmd) {
            cmd.execute();
            this.history.push(cmd);
        }
    }
}

const editor = new TextEditor();
const manager = new CommandManager();

manager.execute(new AppendCommand(editor, "Hello "));
manager.execute(new AppendCommand(editor, "World"));
manager.execute(new AppendCommand(editor, "!"));

console.log(editor.getText());

manager.undo();
console.log(editor.getText());

manager.undo();
console.log(editor.getText());

manager.redo();
console.log(editor.getText());

manager.execute(new AppendCommand(editor, " TypeScript"));
console.log(editor.getText());
