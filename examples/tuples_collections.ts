// Tuples, arrays, destructuring, and functional pipeline methods

type Coordinate = [latitude: number, longitude: number];

const locations: Coordinate[] = [
  [10.8231, 106.6297], // Ho Chi Minh City
  [21.0285, 105.8542], // Hanoi
  [16.0544, 108.2022], // Da Nang
];

const latitudes: number[] = locations.map(([lat, _]) => lat);
const northernLocations: Coordinate[] = locations.filter(([lat, _]) => lat > 15.0);

const averageLat: number =
  latitudes.reduce((acc, lat) => acc + lat, 0) / latitudes.length;

console.log("Average Latitude:", averageLat);
console.log("Northern Locations (> 15°N):", northernLocations);
