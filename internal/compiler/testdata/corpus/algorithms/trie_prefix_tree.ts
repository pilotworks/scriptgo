// @expect: true
// @expect: false
// @expect: true
// @expect: false
class TrieNode {
    children: Map<string, TrieNode> = new Map<string, TrieNode>();
    isEndOfWord: boolean = false;
}

class Trie {
    private root: TrieNode = new TrieNode();

    insert(word: string): void {
        let current = this.root;
        for (let i = 0; i < word.length; i++) {
            const ch = word[i];
            let child = current.children.get(ch);
            if (!child) {
                child = new TrieNode();
                current.children.set(ch, child);
            }
            current = child;
        }
        current.isEndOfWord = true;
    }

    search(word: string): boolean {
        let current = this.root;
        for (let i = 0; i < word.length; i++) {
            const ch = word[i];
            const child = current.children.get(ch);
            if (!child) return false;
            current = child;
        }
        return current.isEndOfWord;
    }

    startsWith(prefix: string): boolean {
        let current = this.root;
        for (let i = 0; i < prefix.length; i++) {
            const ch = prefix[i];
            const child = current.children.get(ch);
            if (!child) return false;
            current = child;
        }
        return true;
    }
}

const trie = new Trie();
trie.insert("apple");
trie.insert("app");
console.log(trie.search("apple"));
console.log(trie.search("appl"));
console.log(trie.startsWith("app"));
console.log(trie.startsWith("ban"));
