interface Address {
  city: string;
  zip: number;
}

interface Company {
  name: string;
  address: Address;
}

interface Employee {
  id: number;
  name: string;
  company: Company;
}

const emp: Employee = {
  id: 1,
  name: "Bob",
  company: {
    name: "Tech Corp",
    address: {
      city: "San Francisco",
      zip: 94105,
    },
  },
};

console.log(JSON.stringify(emp.company.address));
console.log(JSON.stringify(emp));
