// @expect: User #42: Ada Lovelace (18) | Email: ada@computing.org, Backup: none@domain.com | Phone: +44-123-456, Alt: N/A | Tag: math, Remaining Tags: algorithms,pioneer
interface UserProfile {
    id: number;
    info: {
        personal: {
            firstName: string;
            lastName: string;
            age?: number;
        };
        contact: {
            emails: string[];
            phones: [string, string?];
        };
    };
    tags: string[];
}

function processProfile(user: UserProfile): string {
    const {
        id: userId,
        info: {
            personal: { firstName, lastName, age = 18 },
            contact: {
                emails: [primaryEmail, backupEmail = "none@domain.com"],
                phones: [mainPhone, altPhone = "N/A"]
            }
        },
        tags: [firstTag, ...otherTags]
    } = user;

    return `User #${userId}: ${firstName} ${lastName} (${age}) | Email: ${primaryEmail}, Backup: ${backupEmail} | Phone: ${mainPhone}, Alt: ${altPhone} | Tag: ${firstTag}, Remaining Tags: ${otherTags.join(",")}`;
}

const u1: UserProfile = {
    id: 42,
    info: {
        personal: {
            firstName: "Ada",
            lastName: "Lovelace"
        },
        contact: {
            emails: ["ada@computing.org"],
            phones: ["+44-123-456"]
        }
    },
    tags: ["math", "algorithms", "pioneer"]
};

console.log(processProfile(u1));
