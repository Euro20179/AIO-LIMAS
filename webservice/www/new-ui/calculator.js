const TT = {
    Num: "NUM",
    Word: "WORD",
    String: "STRING",
    Add: "+",
    Sub: "-",
    Mul: "*",
    Div: "/",
    Lparen: "(",
    Rparen: ")",
    Comma: ",",
    Semi: ";",
    Eq: "=",
}

const keywords = [
    "var"
]

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

class StringNode extends NodePar {
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

class VarDefNode extends NodePar {
    /**
     * @param {Token} name
     * @param {NodePar} expr
     */
    constructor(name, expr) {
        super()
        this.name = name
        this.expr = expr
    }
}

class FuncDefNode extends NodePar {
    /**
     * @param {Token} name
     * @param {ProgramNode} program
     */
    constructor(name, program) {
        super()
        this.name = name
        this.program = program
        this.closure = new SymbolTable()
    }
}


class CallNode extends NodePar {
    /**
     * @param {Token} name
     * @param {NodePar[]} inner
     */
    constructor(name, inner) {
        super()
        this.name = name
        this.inner = inner
    }
}

class ExprNode extends NodePar {
    /**
     * @param {NodePar} value
     */
    constructor(value) {
        super()
        this.value = value
    }
}

class ProgramNode extends NodePar { }

/**
* @param {string} input
* @param {SymbolTable} symbols
*/
function parseExpression(input, symbols) {
    const tokens = lex(input);
    let parser = new Parser(tokens)
    const tree = parser.ast()
    let int = new Interpreter(tree, symbols)
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
    while (text[++curPos]?.match(/[a-zA-Z0-9_]/)) {
        word += text[curPos]
    }
    return [word, curPos]
}

/**
 * @param {string} text
 * @param {number} curPos
 * @returns {[string, number]}
 */
function buildString(text, curPos) {
    let str = text[curPos]
    let escaped = false
    for (let i = curPos + 1; i < text.length; i++) {
        let ch = text[i]
        if (ch === "\\") {
            escaped = true
            continue
        }
        if (escaped) {
            switch (ch) {
                case 'n': ch = "\n"; break
                case 't': ch = "\t"; break
                case '\\': ch = '\\'; break
                default: ch = "\\" + ch
            }
        } else if (ch === '"') {
            break
        }
        str += ch
    }
    return [str, curPos + str.length + 1]
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
        else if (ch.match(/[a-zA-Z_]/)) {
            let [word, newPos] = buildWord(input, pos)
            tokens.push(new Token("Word", word))
            pos = newPos
        }
        //ignore whitespace
        else if (!ch.trim()) {
            pos++
            continue
        }
        else if (ch === '"') {
            pos++
            let [str, newPos] = buildString(input, pos)
            pos = newPos
            tokens.push(new Token("String", str))
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
            if (this.curTok()?.ty === "Lparen") {
                this.back() //let funcCall get the name
                return this.funcCall()
            }
            return new WordNode(tok.value)
        } else if (tok.ty === "String") {
            return new StringNode(tok.value)
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

    funcCall() {
        let name = this.curTok()
        this.next() //skip Lparen
        this.next()
        if (this.curTok().ty === 'Rparen') {
            this.next()
            return new CallNode(name, [])
        }
        let inner = [this.ast_expr()]
        while (this.curTok()?.ty === "Comma") {
            this.next()
            inner.push(this.ast_expr())
        }
        let node = new CallNode(name, inner)
        if (this.curTok()?.ty !== 'Rparen') {
            console.error(`Invalid token: ${this.curTok().ty}, expected ")"`)
            return new NumNode(0)
        }
        this.next() //skip ")"
        return node
    }

    /**
     * @returns {NodePar}
     */
    varDef() {
        this.next() // skip "var"

        if(this.curTok().ty === "Lparen") {
            return this.funcDef()
        }

        let name = this.curTok()

        this.next()
        if(this.curTok()?.ty !== "Eq") {
            console.error("Expected '='")
            return new NumNode(0)
        }
        this.next()
        return new VarDefNode(name, this.ast_expr())
    }

    funcDef() {
        this.next() //skip "("
        this.next() //skip ")"

        let name = this.curTok()
        this.next()
        if(this.curTok()?.ty !== "Eq") {
            console.error("Expected '='")
            return new NumNode(0)
        }
        this.next()
        let program = this.program()
        if(this.curTok().ty !== "Word" || this.curTok().value !== "rav") {
            console.error("Expected 'rav'")
            return new NumNode(0)
        }
        this.next()
        return new FuncDefNode(name, program)
    }

    ast_expr() {
        let t = this.curTok()
        if(t.ty === "Word" && t.value === "var") {
            return this.varDef()
        }
        let expr = new ExprNode(this.arithmatic())
        return expr
    }

    program() {
        let program = new ProgramNode()
        program.addChild(this.ast_expr())
        while (this.curTok()?.ty === "Semi") {
            this.next()
            program.addChild(this.ast_expr())
        }
        return program
    }

    ast() {
        let root = new NodePar()
        let node = this.program()
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

    jsStr() {
        return String(this.jsValue)
    }

    toStr() {
        return new Str(this.jsStr())
    }

    toNum() {
        return new Num(Number(this.jsValue))
    }

    /**
     * @param {Type} right
     */
    add(right) {
        console.error(`Unable to add ${this.constructor.name} and ${right.constructor.name}`)
        return right
    }
    /**
     * @param {Type} right
     */
    sub(right) {
        console.error(`Unable to add ${this.constructor.name} and ${right.constructor.name}`)
        return right
    }
    /**
     * @param {Type} right
     */
    mul(right) {
        console.error(`Unable to add ${this.constructor.name} and ${right.constructor.name}`)
        return right
    }
    /**
     * @param {Type} right
     */
    div(right) {
        console.error(`Unable to add ${this.constructor.name} and ${right.constructor.name}`)
        return right
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

    /**
     * @param {Type[]} params
     */
    call(params) {
        console.error(`Cannot call ${this.constructor.name}`)
    }
}

class Func extends Type {
    /**
     * @param {Type[]} params
     */
    call(params) {
        return this.jsValue(...params)
    }
}

class Num extends Type {
    /**
     * @param {Type} right
     */
    add(right) {
        return new Num(this.jsValue + right.toNum().jsValue)
    }

    /**
     * @param {Type} right
     */
    sub(right) {
        return new Num(this.jsValue - right.toNum().jsValue)
    }

    /**
     * @param {Type} right
     */
    mul(right) {
        return new Num(this.jsValue * right.toNum().jsValue)
    }

    /**
     * @param {Type} right
     */
    div(right) {
        return new Num(this.jsValue / right.toNum().jsValue)
    }

    /**
     * @param {Type} right
     */
    lt(right) {
        if (this.jsValue < Number(right.jsValue)) {
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

    toNum() {
        return this
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
        this.jsValue = this.jsValue.replaceAll(String(right.jsValue), "")
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

class List extends Type {

    jsStr() {
        let str = ""
        for (let item of this.jsValue) {
            str += item.jsStr() + ","
        }
        return str.slice(0, -1)
    }

}

class SymbolTable {
    constructor() {
        this.symbols = new Map()

        this.setupDefaultFunctions()
    }
    setupDefaultFunctions() {
        //@ts-ignore
        this.symbols.set("abs", new Func(n => {
            return new Num(Math.abs(n.toNum().jsValue))
        }))

        this.symbols.set("str", new Func(n => {
            return n.toStr()
        }))

        this.symbols.set("max", new Func((...items) => {
            let max = items[0].toNum()
            for(let item of items) {
                if (item.toNum().jsValue > max.jsValue) {
                    max = item
                }
            }
            return max
        }))

        this.symbols.set("min", new Func((...items) => {
            let min = items[0].toNum()
            for(let item of items) {
                if(item.toNum().jsValue < min.jsValue) {
                    min = item
                }
            }
            return min
        }))
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

    copy() {
        let copy = new SymbolTable()
        for(let item of this.symbols.entries()) {
            copy.set(item[0], item[1])
        }
        return copy
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
    StringNode(node) {
        return new Str(node.value)
    }

    /**
    * @param {WordNode} node
    */
    WordNode(node) {
        if (this.symbolTable.get(node.value)) {
            return this.symbolTable.get(node.value)
        }
        return new Str(node.value)
    }

    /**
     * @param {UnOpNode} node
     */
    UnOpNode(node) {
        let right = this.interpretNode(node.right)
        if (node.operator.ty === "Add") {
            return right
        } else {
            return right.mul(new Num(-1))
        }
    }

    /**
     * @param {BinOpNode} node
     */
    BinOpNode(node) {
        let left = this.interpretNode(node.left)
        let right = this.interpretNode(node.right)

        if (node.operator.ty === "Add") {
            return left.add(right)
        } else if (node.operator.ty === "Sub") {
            return left.sub(right)
        } else if (node.operator.ty === "Mul") {
            return left.mul(right)
        } else if (node.operator.ty === "Div") {
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
     * @param {CallNode} node
     *
     * @returns {Type}
     */
    CallNode(node) {
        let inner = node.inner.map(v => this.interpretNode(v))
        let name = node.name
        let value = this.symbolTable.get(name.value)
        if (!value) {
            return new Num(0)
        }
        return value.call(inner)
    }

    /**
     * @param {VarDefNode} node
     */
    VarDefNode(node) {
        let val = this.interpretNode(node.expr)
        this.symbolTable.set(node.name.value, val)
        return val
    }

    /**
     * @param {FuncDefNode} node
     */
    FuncDefNode(node) {
        node.closure = this.symbolTable.copy()

        this.symbolTable.set(node.name.value, new Func((...items) => {
            for(let i = 0; i < items.length; i++) {
                node.closure.set(`arg${i}`, items[i])
            }
            let interpreter = new Interpreter(node.program, node.closure)
            return interpreter.interpretNode(node.program)
        }))

        return this.symbolTable.get(node.name.value)
    }

    /**
     * @param {ExprNode} node
     */
    ExprNode(node) {
        return this.interpretNode(node.value)
    }

    /**
    * @param {ProgramNode} node
    */
    ProgramNode(node) {
        let values = []
        for (let child of node.children) {
            if (!(child.constructor.name in this)) {
                console.error(`Unimplmemented: ${child.constructor.name}`)
            } else {
                values.push(this.interpretNode(child))
            }
        }
        return values.slice(-1)[0]
    }

    interpret() {
        return this.interpretNode(this.tree)
    }
}
