const TT = {
    Num: "NUM",
    Word: "WORD",
    Add: "+",
    Sub: "-",
    Mul: "*",
    Div: "/",
    Lparen: "(",
    Rparen: ")",
}

class Token {
    /**
     * @param {keyof TT} ty
     * @param {string} value
     */
    constructor(ty, value) {
        this.ty = ty
        this.value = value
    }
}

// AST node types
class NodePar {
    /**
    *@param {NodePar?} child
    */
    constructor(child = null) {
        /**@type {NodePar[]}*/
        this.children = []
        if (child) {
            this.children.push(child)
        }
    }

    /**
     * @param {NodePar} child
     */
    addChild(child) {
        this.children.push(child)
    }
}

class NumNode extends NodePar {
    /**
    * @param {number} value
    */
    constructor(value) {
        super()
        this.value = value;
    }
}

class WordNode extends NodePar {
    /**
     * @param {string} value
     */
    constructor(value) {
        super()
        this.value = value
    }
}

class ErrorNode extends NodePar {
    /**
     * @param {string} value
     */
    constructor(value) {
        super()
        this.value = value
    }
}

class UnOpNode extends NodePar {
    /**
     * @param {NodePar} right
     * @param {Token} operator
     */
    constructor(right, operator) {
        super()
        this.right = right
        this.operator = operator
    }
}

class BinOpNode extends NodePar {
    /**
    * @param {NodePar} left
    * @param {Token} operator
    * @param {NodePar} right
    */
    constructor(left, operator, right) {
        super(); // Use the first token type as default
        this.left = left;
        this.operator = operator;
        this.right = right;
    }
}

class ExprNode extends NodePar { }

/**
* @param {string} input
*/
function parseExpression(input) {
    const tokens = lex(input);
    let parser = new Parser(tokens)
    const tree = parser.ast()
    let int = new Interpreter(tree, new SymbolTable())
    let values = int.interpret()
    return values
    // return ast(tokens);
}

/**
 * @param {string} text
 * @param {number} curPos
 * @returns {[string, number]}
 */
function buildNumber(text, curPos) {
    let num = text[curPos]
    while ("0123456789".includes(text[++curPos])) {
        num += text[curPos]
    }
    return [num, curPos]
}

/**
 * @param {string} text
 * @param {number} curPos
 * @returns {[string, number]}
 */
function buildWord(text, curPos) {
    let word = text[curPos]
    while (text[++curPos]?.match(/[a-zA-Z0-9]/)) {
        word += text[curPos]
    }
    return [word, curPos]
}

/**
* @param {string} input
* @returns {Token[]}
*/
function lex(input) {
    let pos = 0;

    let tokens = [];

    while (pos < input.length) {
        let ch = input[pos]
        if ("0123456789".includes(ch)) {
            let [num, newPos] = buildNumber(input, pos)
            tokens.push(new Token("Num", num));
            pos = newPos
        }
        else if (ch.match(/[a-zA-Z]/)) {
            let [word, newPos] = buildWord(input, pos)
            tokens.push(new Token("Word", word))
            pos = newPos
        }
        //ignore whitespace
        else if (!ch.trim()) {
            pos++
            continue
        }
        else {
            let foundTok = false
            /**@type {keyof TT}*/
            let tok
            for (tok in TT) {
                if (ch === TT[tok]) {
                    tokens.push(new Token(tok, ch))
                    pos++
                    foundTok = true
                    break
                }
            }
            if (!foundTok) {
                console.error(`Invalid token: ${ch}`)
                pos++
            }
        }
    }

    return tokens;
}

class Parser {
    /**
     * @param {Token[]} tokens
     */
    constructor(tokens) {
        this.i = 0
        this.tokens = tokens
    }

    next() {
        this.i++
        return this.i < this.tokens.length
    }

    back() {
        this.i--
    }

    curTok() {
        return this.tokens[this.i]
    }
    /**
     * @returns {NodePar}
     */
    atom() {
        let tok = this.curTok()
        if (!tok) return new ErrorNode("Ran out of tokens")

        this.next()

        if (tok.ty === "Num") {
            return new NumNode(Number(tok.value))
        } else if (tok.ty === "Word") {
            return new WordNode(tok.value)
        } else if (tok.ty === "Lparen") {
            let node = this.ast_expr()
            this.next() //skip Rparen
            return node
        }
        return new ErrorNode(`Invalid token: (${tok.ty} ${tok.value})`)
    }

    signedAtom() {
        let tok = this.curTok()
        if ("+-".includes(tok.value)) {
            this.next()
            let right = this.atom()
            return new UnOpNode(right, tok)
        }
        return this.atom()
    }

    /**
     * @returns {NodePar}
     */
    product() {
        let left = this.signedAtom()
        let op = this.curTok()
        while (op && "*/".includes(op.value)) {
            this.next()
            let right = this.product()
            left = new BinOpNode(left, op, right)
            op = this.curTok()
        }
        return left
    }

    arithmatic() {
        let left = this.product()
        let op = this.curTok()
        while (op && "+-".includes(op.value)) {
            this.next()
            let right = this.arithmatic()
            left = new BinOpNode(left, op, right)
            op = this.curTok()
        }
        return left
    }

    ast_expr() {
        let expr = new ExprNode()

        expr.addChild(this.arithmatic())

        return expr
    }

    ast() {
        let root = new NodePar()
        let node = this.ast_expr()
        root.addChild(node)

        return node;
    }
}

class Type {
    /**
     * @param {any} jsValue
     */
    constructor(jsValue) {
        this.jsValue = jsValue
    }

    /**
     * @param {Type} right
     */
    add(right) {
        console.error(`Unable to add ${this.constructor.name} and ${right.constructor.name}`)
        return this
    }
    /**
     * @param {Type} right
     */
    sub(right) {
        console.error(`Unable to add ${this.constructor.name} and ${right.constructor.name}`)
        return this
    }
    /**
     * @param {Type} right
     */
    mul(right) {
        console.error(`Unable to add ${this.constructor.name} and ${right.constructor.name}`)
        return this
    }
    /**
     * @param {Type} right
     */
    div(right) {
        console.error(`Unable to add ${this.constructor.name} and ${right.constructor.name}`)
        return this
    }

    /**
     * @param {Type} right
     */
    lt(right) {
        console.error(`Unable to compare ${this.constructor.name} < ${right.constructor.name}`)
        return false
    }

    /**
     * @param {Type} right
     */
    gt(right) {
        console.error(`Unable to compare ${this.constructor.name} > ${right.constructor.name}`)
        return false
    }
}

class Num extends Type {
    /**
     * @param {Type} right
     */
    add(right) {
        this.jsValue += Number(right.jsValue)
        return this
    }

    /**
     * @param {Type} right
     */
    sub(right) {
        this.jsValue -= Number(right.jsValue)
        return this
    }

    /**
     * @param {Type} right
     */
    mul(right) {
        this.jsValue *= Number(right.jsValue)
        return this
    }

    /**
     * @param {Type} right
     */
    div(right) {
        this.jsValue /= Number(right.jsValue)
        return this
    }

    /**
     * @param {Type} right
     */
    lt(right) {
        if(this.jsValue < Number(right.jsValue)) {
            return true
        }
        return false
    }

    /**
     * @param {Type} right
     */
    gt(right) {
        return this.jsValue > Number(right.jsValue)
    }    
}

class Str extends Type {
    /**
     * @param {Type} right
     */
    add(right) {
        this.jsValue += String(right.jsValue)
        return this
    }
    /**
     * @param {Type} right
     */
    sub(right) {
        this.jsValue = this.jsValue.replaceAll(String(right.jsValue))
        return this
    }    
    /**
     * @param {Type} right
     */
    mul(right) {
        this.jsValue = this.jsValue.repeat(Number(right.jsValue))
        return this
    }    
}

class SymbolTable {
    constructor() {
        this.symbols = new Map()
    }
    /**
     * @param {string} name
     * @param {Type} value
     */
    set(name, value) {
        this.symbols.set(name, value)
    }
    /**
     * @param {string} name
     */
    get(name) {
        return this.symbols.get(name)
    }
}

class Interpreter {
    /**
    * @param {NodePar} tree
    * @param {SymbolTable} symbolTable 
    */
    constructor(tree, symbolTable) {
        this.tree = tree
        this.symbolTable = symbolTable
    }

    /**
    * @param {NumNode} node
    */
    NumNode(node) {
        return new Num(node.value)
    }

    /**
    * @param {WordNode} node
    */
    WordNode(node) {
        if (this.symbolTable.get(node.value)) {
            return this.symbolTable.get(node.value)
        }
        return node.value
    }

    /**
     * @param {UnOpNode} node
     */
    UnOpNode(node) {
        let right = this.interpretNode(node.right)
        if(node.operator.ty === "Add") {
            return right
        } else {
            return right.mul(new Num(-1))
        }
    }

    /**
     * @param {BinOpNode} node
     */
    BinOpNode(node) {
        let right = this.interpretNode(node.right)
        let left = this.interpretNode(node.left)

        if(node.operator.ty === "Add") {
            return left.add(right)
        } else if(node.operator.ty === "Sub") {
            return left.sub(right)
        } else if(node.operator.ty === "Mul") {
            return left.mul(right)
        } else if(node.operator.ty === "Div") {
            return left.div(right)
        }
        return right
    }

    /**
     * @param {NodePar} node
     *
     * @returns {Type}
     */
    interpretNode(node) {
        //@ts-ignore
        return this[node.constructor.name](node)
    }

    /**
     * @param {ExprNode} node
     */
    ExprNode(node) {
        for(let child of node.children) {
            if(!(child.constructor.name in this)) {
                console.error(`Unimplmemented: ${child.constructor.name}`)
            } else {
                return this.interpretNode(child)
            }
        }
        return new Num(0)
    }

    interpret() {
        return this.ExprNode(this.tree)
    }
}
