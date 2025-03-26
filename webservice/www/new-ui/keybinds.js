
const nop = () => { }
class KeyTree {
    /**
     * @param {string} key
     * @param {Function} onExec
     */
    constructor(key, onExec) {
        this.key = key
        /**@type {Map<string, KeyTree>}*/
        this.next = new Map

        this.onExec = onExec
    }

    /**
     * @param {string[]} keyseq
     * @param {Function} onExec
     */
    add(keyseq, onExec) {
        /**@type {KeyTree}*/
        let current = this
        for (let i = 0; i < keyseq.length - 1; i++) {
            const key = keyseq[i]
            if (current.next.has(key)) {
                current = /**@type {KeyTree} */ (current.next.get(key))
            } else {
                const newTree = new KeyTree(key, nop)
                current.next.set(key, newTree)
                current = newTree
            }
        }

        const key = keyseq[keyseq.length - 1]
        if (current.next.get(key)) {
            /**@type {KeyTree}*/
            (current.next.get(key)).onExec = onExec
        } else {
            current.next.set(key, new KeyTree(key, onExec))
        }
    }

    /**
     * @param {string} key
     * @returns {boolean}
     */
    has(key) {
        return this.next.has(key)
    }

    /**
     * @param {string} key
     * @returns {KeyTree | undefined}
     */
    follow(key) {
        return this.next.get(key)
    }

    isLeaf() {
        return this.next.size === 0
    }

    exec() {
        this.onExec()
    }
}

class KeyState {
    /**
     * @param {KeyTree} tree
     */
    constructor(tree) {
        this.tree = tree

        this.currentPos = this.tree

        /**@type {number?} */
        this.timeout = null
    }

    /**
     * @param {string} key
     *
     * @returns {boolean} Whether or not the key executed something
     */
    applyKey(key) {

        if (this.timeout) {
            clearTimeout(this.timeout)
        } else {
            this.timeout = setTimeout(this.execNow.bind(this), 1000)
        }

        let next = this.currentPos.follow(key)
        if (next) {
            //the user hit the end of the key tree, we should just execute, and reset state
            if (next.isLeaf()) {
                this.execNow(next)
                return true
            } else {
                this.currentPos = next
            }
        } else {
            //if there is no next, we need to reset the tree state because the user has started a new sequence of key presses
            this.resetState()
        }

        return false
    }

    /**
     * @param {KeyTree?} node Leave null to execute the current node
     */
    execNow(node = null) {
        node ||= this.currentPos
        console.log(node)
        node.exec()
        this.resetState()
    }

    resetState() {
        if(this.timeout)
            clearTimeout(this.timeout)
        this.timeout = null
        this.currentPos = this.tree
    }
}

const searchBar = /**@type {HTMLInputElement} */(document.querySelector("[name=\"search-query\"]"))

let KEY_TREE = new KeyTree("", nop)
KEY_TREE.add(["g", "i"], searchBar.focus.bind(searchBar))
KEY_TREE.add(["Escape"], () => {
    if(document.activeElement?.tagName === "INPUT") {
        document.activeElement.blur()
    }
})

const ks = new KeyState(KEY_TREE)

/**
* @param {string} key
*/
function isModifier(key) {
    return key === "Control" || key === "Shift" || key === "Alt" || key === "Meta"
}

document.addEventListener("keydown", e => {
    console.log(e.key)
    //if the user is interacting with input elements, just let them do their thing
    //if it's a modifier we should drop it
    //TODO: add some way for the keystate to keep track of modifiers
    if ((e.key !== "Escape" && document.activeElement?.tagName === "INPUT") || isModifier(e.key)) {
        return
    }
    if (ks.applyKey(e.key)) {
        e.preventDefault()
    }
})

