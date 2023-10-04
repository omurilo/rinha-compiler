const fs = require("fs");

function interpret(env, node) {
	switch (node.kind) {
		case "Int":
			return node.value;
		case "Str":
			return node.value;
		case "Bool":
			return !!node.value;
		case "Print": {
			const value = interpret(env, node.value);
			if (value?.node?.kind === "Function") {
				console.log("<#closure>");
			} else if (Array.isArray(value)) {
				console.log(`(${value[0]}, ${value[1]})`);
			} else {
				console.log(value);
			}

			return value;
		}
		case "Call": {
			const { node: calleeNode, env: calleeEnvironment } = interpret(env, node.callee);
			const args = node.arguments.map((arg) => interpret(env, arg));
			const newEnvironment = { ...calleeEnvironment };
			calleeNode.parameters.forEach((param, index) => {
				newEnvironment[param.text] = args[index];
			});
			return interpret(newEnvironment, calleeNode.value);
		}
		case "Function":
			return { env, node };
		case "Let": {
			env[node.name.text] = interpret(env, node.value);
			return interpret({ ...env }, node.next);
		}
		case "Var": {
			const newVar = node.text;
			if (newVar in env) {
				return env[newVar];
			} else {
				return console.error(`Variable ${newVar} not defined`);
			}
		}
		case "Tuple":
			return [interpret(env, node.first), interpret(env, node.second)];
		case "First": {
			const value = interpret(env, node.value);
			if (Array.isArray(value) && value.length == 2) {
				return value[0];
			}

			return console.error("Expected a tuple for first operation");
		}
		case "Second": {
			const value = interpret(env, node.value);
			if (Array.isArray(value) && value.length == 2) {
				return value[1];
			}

			return console.error("Expected a tuple for second operation");
		}
		case "If": {
			const condition = interpret(env, node.condition);
			if (!!condition) {
				return interpret(env, node.then);
			}

			return interpret(env, node.otherwise);
		}
		case "Binary": {
			const {op} = node
      const lhs = interpret(env, node.lhs)
      const rhs = interpret(env, node.rhs)

			switch (op) {
				case "Add": {
          if (typeof lhs === "string" || typeof rhs === "string") {
            return `${lhs}${rhs}`
          }
    
          if (typeof lhs === "number" && typeof rhs === "number") {
            return lhs + rhs
          }
    
          Error(binaryValue.Location, "invalid add operation")
        }
				case "Sub":
          if (typeof lhs !== "number" || typeof rhs !== "number") {
            throw new Error("Invalid Sub operation")
          }
					return lhs - rhs;
				case "Mul":
          if (typeof lhs !== "number" || typeof rhs !== "number") {
            throw new Error("Invalid Mul operation")
          }
					return lhs * rhs;
				case "Div": {
          if (typeof lhs !== "number" || typeof rhs !== "number") {
            throw new Error("Invalid Div operation")
          }

					if (rhs == 0) {
						throw new Error("division by 0 not expected");
					}

					return parseInt(lhs / rhs, 10);
				}
				case "Eq":
					return lhs === rhs;
				case "Neq":
					return lhs !== rhs;
				case "Lt":
          if (typeof lhs !== "number" || typeof rhs !== "number") {
            throw new Error("Invalid Lt operation")
          }
					return lhs < rhs;
				case "Lte":
          if (typeof lhs !== "number" || typeof rhs !== "number") {
            throw new Error("Invalid Lte operation")
          }
					return lhs <= rhs;
				case "Gt":
          if (typeof lhs !== "number" || typeof rhs !== "number") {
            throw new Error("Invalid Gt operation")
          }
					return lhs > rhs;
				case "Gte":
          if (typeof lhs !== "number" || typeof rhs !== "number") {
            throw new Error("Invalid Gte operation")
          }
					return lhs >= rhs;
				case "And":
					return !!lhs && !!rhs;
				case "Or":
					return !!lhs || !!rhs;
				case "Rem":
          if (typeof lhs !== "number" || typeof rhs !== "number") {
            throw new Error("Invalid Rem operation")
          }
					return lhs % rhs;
			}
		}
		default:
			break;
	}
}

module.exports = function () {
  try {
    const DEFAULT_FILE_PATH = "/var/rinha/source.rinha.json"
		const rawData = fs.readFileSync(process.argv[2] || DEFAULT_FILE_PATH);
		const ast = JSON.parse(rawData);
		const environment = {};
		return interpret(environment, ast.expression);
	} catch (error) {
		console.error("Erro ao executar o código da rinha:", error.message);
		console.error(error);
		return null;
	}
}()
