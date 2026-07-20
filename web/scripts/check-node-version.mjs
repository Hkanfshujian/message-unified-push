const minimum = [20, 19, 0]
const current = process.versions.node.split('.').map(Number)

const ok = current[0] > minimum[0]
  || (current[0] === minimum[0] && current[1] > minimum[1])
  || (current[0] === minimum[0] && current[1] === minimum[1] && current[2] >= minimum[2])

if (!ok) {
  console.error(`Node.js ${process.versions.node} is not supported. Please use Node.js >=20.19.0.`)
  process.exit(1)
}
