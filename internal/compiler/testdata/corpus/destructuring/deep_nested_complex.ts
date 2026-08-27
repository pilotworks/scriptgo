// @expect: Title: System Design
// @expect: Author: Bob
// @expect: Rating: 4.5
// @expect: First tag: tech
// @expect: Page count: 350
type Book = {
    title: string;
    details: {
        author: {
            name: string;
        };
        stats: {
            rating: number;
            pages?: number;
        };
        tags: string[];
    };
};

const book: Book = {
    title: "System Design",
    details: {
        author: { name: "Bob" },
        stats: { rating: 4.5 },
        tags: ["tech", "architecture"]
    }
};

const {
    title,
    details: {
        author: { name: authorName },
        stats: { rating, pages = 350 },
        tags: [firstTag]
    }
} = book;

console.log("Title: " + title);
console.log("Author: " + authorName);
console.log("Rating: " + rating);
console.log("First tag: " + firstTag);
console.log("Page count: " + pages);
