package MCTS

import (
	"log"
	"math/rand"
	"sort"
)

type treeNode struct {
	parent         *treeNode   // What node contains this node? Root node's parent is nil.
	move           Move        // What move lead to this node? Root node's action is nil.
	state          GameState   // What is the game state at this node?
	totalOutcome   float64     // What is the sum of all outcomes computed for this node and its children? From the point of view of a single player.
	probability    float64     // The probability of this move even occurring (when it involves randomness) as 0.0-1.0.
	visits         uint64      // How many times has this node been studied? Used with totalValue to compute an average value for the node.
	untriedMoves   []Move      // What moves have not yet been explored from this state?
	children       []*treeNode // The children of this node, can be many.
	ucbC           float64     // The UCB constant used in selection calculations.
	selectionScore float64     // The computed score for this node used in selection, balanced between exploitation and exploration.
}

//newTreeNode creates a new well-formed tree node
func newTreeNode(parent *treeNode, move Move, state GameState, ucbC float64) *treeNode {
	//Sanity check the move probability
	var probability = move.Probability()
	if probability < 0.0 || probability > 1.0 {
		log.Panicf("Move cannot have a probability outside of the range 0.0-1.0:%f", probability)
	}
	//Construct the new node
	var node = treeNode{
		parent:         parent,
		move:           move,
		state:          state,
		totalOutcome:   0.0,                    //No outcome yet
		probability:    probability,            //Some moves happen so rarely we want to weight the value of their influence
		visits:         0,                      //No visit yet
		untriedMoves:   state.AvailableMoves(), //Initially the node starts with every node unexplored
		children:       nil,                    //not children yet
		ucbC:           ucbC,                   //Whole three uses same constant
		selectionScore: 0,                      //No value yet
	}
	//We're working with pointers
	return &node
}

//getVisits returns the visit to a node,0 if the node doesn't exist(for when a root checks its parent)
func (n *treeNode) getVisits() uint64 {
	if n == nil {
		return 0
	}
	return n.visits
}

//computeSelectionScore prepares the selection score of a single child
func (n *treeNode) computeSelectionScore() {
	n.selectionScore = upperConfidenceBound(n.totalOutcome, n.ucbC, n.parent.getVisits(), n.visits)
}

//selectChild picks the child with the highest selection score(Balancing exploration and exploitation)
func (n *treeNode) selectChild() *treeNode {
	//Sort the children by their UCB,balance winning children with unexplored children.
	sort.Sort(bySelectionScore(n.children))
	return n.children[0]
}

//addOutcome adds the outcome value from a computation involving the node or one of its children
//Every outcome value in the tree is from the perspective of a particular player.Higher outcomes mean better
//Winning situation for the player
func (n *treeNode) addOutcome(outcome float64) {
	//Allow the root to call this on its parent with no ill effect
	if n != nil {
		//Some nodes are so unlikely to be visited ,the outcome should be weighted
		var weightedOutcome = outcome * n.probability
		//Update this bode's data
		n.totalOutcome += weightedOutcome
		n.visits++
		//pass the value up to the parent as well
		n.parent.addOutcome(weightedOutcome) //Will recurse up the tree to the root
		//Now that the parent is also updated
		n.computeSelectionScore()
	}
}

//makeRandomUntriedMove makes a random untried move and builds another node in the tree form the result
func (n *treeNode) makeRandomUntriedMove() *treeNode {
	//Select a random move we haven't tried
	var i = rand.Intn(len(n.untriedMoves))
	var move = n.untriedMoves[i]
	//Remove it from the untried moves
	n.untriedMoves = append(n.untriedMoves[:i], n.untriedMoves[i+1:]...)
	//Clone the node's state so we don't alter it
	var newState = n.state.Clone()
	newState.MakeMove(move)

	//Build more of the tree
	var child = newTreeNode(n, move, newState, n.ucbC)
	n.children = append(n.children, child) //End of children list are the children with lowest selection scores.(e.g.no visit)
	//return a game state that can be used for simulations
	return child
}

//bySelectionScore implements sort.Interface to sort *descending* by selection score
//Example:sort.Sort(bySelectionScore(nodes))

type bySelectionScore []*treeNode

func (b bySelectionScore) Len() int {
	return len(b)
}

func (b bySelectionScore) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}
func (b bySelectionScore) Less(i, j int) bool {
	return b[i].visits > b[j].visits
}

// byVisits implements sort.Interface to sort *descending* by visits.
// Example: sort.Sort(byVisits(nodes))
type byVisits []*treeNode

func (a byVisits) Len() int           { return len(a) }
func (a byVisits) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byVisits) Less(i, j int) bool { return a[i].visits > a[j].visits }
