import * as fs from "node:fs";
import { promises } from "node:fs";

async function run(): Promise<void> {
    const dir = "test_fs_promises_tmp";
    const mkdirOpts: fs.MkdirOptions = { recursive: true };
    await promises.mkdir(dir, mkdirOpts);

    const f = dir + "/data.txt";
    await promises.writeFile(f, "Promise FS Works!");
    
    const content = await promises.readFile(f);
    console.log(content);

    const st = await promises.stat(f);
    console.log(st.size);
    console.log(st.isFile());

    const c = dir + "/data_copy.txt";
    await promises.copyFile(f, c);

    const entries = await promises.readdir(dir);
    console.log(entries.length);

    await promises.unlink(f);
    await promises.unlink(c);
    const rmOpts: fs.RmOptions = { recursive: true, force: true };
    await promises.rm(dir, rmOpts);
    console.log(fs.existsSync(dir));
}

run();
