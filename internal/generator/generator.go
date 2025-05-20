package generator

import (
	"context"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/amirahmetzanov/go_project/internal/workerpool"
)

// NamesByLetter contains predefined name lists for each letter of the alphabet
var NamesByLetter = map[string][]string{
	"A": {"Adam", "Anna", "Alex", "Ava", "Andrew", "Alice", "Aaron", "Abigail", "Anthony", "Amelia", "Austin", "Audrey", "Adrian", "Aria", "Alan", "Allison", "Aiden", "Aubrey", "Arthur", "Aurora"},
	"B": {"Benjamin", "Bella", "Brandon", "Bailey", "Brian", "Brooke", "Blake", "Brianna", "Bruce", "Brooklyn", "Bradley", "Brielle", "Brett", "Bethany", "Boris", "Bianca", "Bennett", "Bridget", "Byron", "Beatrice"},
	"C": {"Charles", "Charlotte", "Christopher", "Catherine", "Caleb", "Chloe", "Cameron", "Claire", "Cole", "Caroline", "Colin", "Camila", "Cooper", "Cecilia", "Connor", "Christina", "Calvin", "Cora", "Carter", "Celeste"},
	"D": {"David", "Daisy", "Daniel", "Diana", "Dylan", "Danielle", "Derek", "Destiny", "Dominic", "Delilah", "Diego", "Dahlia", "Damian", "Dorothy", "Dean", "Daphne", "Dustin", "Delaney", "Darius", "Denise"},
	"E": {"Edward", "Emma", "Ethan", "Emily", "Evan", "Elizabeth", "Elijah", "Ella", "Eric", "Eleanor", "Emmett", "Eden", "Elliott", "Evelyn", "Edwin", "Elise", "Edgar", "Eliana", "Ezra", "Eva"},
	"F": {"Frank", "Faith", "Felix", "Fiona", "Finn", "Felicity", "Fernando", "Florence", "Frederick", "Francesca", "Francisco", "Freya", "Fabian", "Flora", "Franklin", "Farrah", "Fletcher", "Fatima", "Forbes", "Finley"},
	"G": {"George", "Grace", "Gabriel", "Gabriella", "Grant", "Georgia", "Gregory", "Gemma", "Gavin", "Genevieve", "Graham", "Giselle", "Garrett", "Gwendolyn", "Glenn", "Gloria", "Gordon", "Gillian", "Gary", "Gianna"},
	"H": {"Henry", "Hannah", "Harrison", "Haley", "Harry", "Harper", "Hudson", "Hailey", "Hunter", "Heather", "Hugo", "Holly", "Harold", "Hope", "Hayes", "Helena", "Hector", "Harlow", "Howard", "Harmony"},
	"I": {"Isaac", "Isabella", "Ian", "Ivy", "Ivan", "Iris", "Isaiah", "Isla", "Ignacio", "Ingrid", "Ismael", "Iliana", "Irvin", "India", "Ibrahim", "Imogen", "Ira", "Irene", "Isiah", "Ines"},
	"J": {"James", "Jasmine", "John", "Julia", "Jacob", "Jennifer", "Joseph", "Jade", "Joshua", "Jessica", "Jack", "Jordan", "Julian", "Josephine", "Jason", "Jane", "Jeremy", "June", "Jesse", "Jocelyn"},
	"K": {"Kevin", "Katherine", "Kenneth", "Kate", "Keith", "Kylie", "Kyle", "Kimberly", "Kai", "Kelly", "Karl", "Kennedy", "Kane", "Kayla", "Keegan", "Kendall", "Kieran", "Keira", "Kaleb", "Kaitlyn"},
	"L": {"Lucas", "Lily", "Logan", "Lucy", "Liam", "Leah", "Leo", "Luna", "Louis", "Layla", "Leonard", "Lauren", "Lawrence", "Lillian", "Lance", "Leslie", "Lewis", "Lola", "Landon", "Lydia"},
	"M": {"Michael", "Maria", "Matthew", "Madison", "Mark", "Mia", "Miles", "Megan", "Maxwell", "Maya", "Morgan", "Molly", "Marcus", "Michelle", "Mason", "Melissa", "Malcolm", "Miranda", "Martin", "Madeline"},
	"N": {"Nathan", "Natalie", "Nicholas", "Nicole", "Noah", "Nina", "Neil", "Nora", "Nolan", "Naomi", "Nelson", "Noelle", "Norman", "Nancy", "Nathaniel", "Natasha", "Noel", "Nevaeh", "Niall", "Nadine"},
	"O": {"Oliver", "Olivia", "Oscar", "Ophelia", "Owen", "Octavia", "Omar", "Odette", "Otto", "Opal", "Orion", "Olga", "Orlando", "Olive", "Otis", "Olympia", "Oswald", "Orla", "Osvaldo", "Oakley"},
	"P": {"Peter", "Penelope", "Paul", "Patricia", "Patrick", "Paige", "Philip", "Phoebe", "Percy", "Paula", "Parker", "Priscilla", "Preston", "Pearl", "Pierce", "Poppy", "Perry", "Paloma", "Princeton", "Paulina"},
	"Q": {"Quentin", "Quinn", "Quincy", "Quiana", "Quinton", "Quimby", "Quade", "Questa", "Quest", "Queenie", "Quigley", "Quilla", "Quillan", "Queen", "Quintessa", "Quimbly", "Quaid", "Querida", "Quirin", "Qadira"},
	"R": {"Robert", "Rachel", "Richard", "Rebecca", "Ryan", "Ruby", "Raymond", "Rose", "Roman", "Renee", "Russell", "Riley", "Reed", "Ruth", "Rhett", "Regina", "Rowan", "Rosalie", "Reginald", "Roxanne"},
	"S": {"Samuel", "Sophia", "Steven", "Sarah", "Simon", "Samantha", "Sebastian", "Stella", "Scott", "Sydney", "Seth", "Savannah", "Shane", "Scarlett", "Spencer", "Sofia", "Stanley", "Serena", "Stefan", "Sienna"},
	"T": {"Thomas", "Taylor", "Timothy", "Tiffany", "Tyler", "Tessa", "Theodore", "Trinity", "Trevor", "Talia", "Travis", "Tara", "Tristan", "Teagan", "Tobias", "Tamara", "Tony", "Thea", "Trent", "Tatiana"},
	"U": {"Ulysses", "Uma", "Uriel", "Ursula", "Urban", "Unity", "Uziel", "Udele", "Upton", "Ula", "Umar", "Umi", "Usher", "Uta", "Urijah", "Ulla", "Urson", "Ulani", "Udo", "Ulyana"},
	"V": {"Victor", "Victoria", "Vincent", "Valerie", "Vaughn", "Vanessa", "Vernon", "Vera", "Vince", "Violet", "Virgil", "Venus", "Valentine", "Vivian", "Van", "Valentina", "Vito", "Veronica", "Vance", "Virginia"},
	"W": {"William", "Wendy", "Walter", "Willow", "Wesley", "Whitney", "Warren", "Willa", "Winston", "Winter", "Wade", "Waverly", "Wayne", "Wallis", "Wyatt", "Wren", "Wallace", "Winona", "Weston", "Wilhelmina"},
	"X": {"Xavier", "Xenia", "Xander", "Xiomara", "Xerxes", "Xyla", "Xzavier", "Xanthe", "Xavi", "Xena", "Xeno", "Ximena", "Xoan", "Xandra", "Xylon", "Xia", "Xuan", "Xaviera", "Xaden", "Xanthia"},
	"Y": {"Yuri", "Yasmine", "Yosef", "Yara", "Yale", "Yvonne", "Yehuda", "Yasmin", "York", "Yolanda", "Yakov", "Yuna", "Yannick", "Yvette", "Yoel", "Yana", "Youssef", "Yesenia", "Yuval", "Yuliana"},
	"Z": {"Zachary", "Zoe", "Zane", "Zelda", "Zeus", "Zara", "Zion", "Zara", "Zack", "Zahara", "Zeke", "Zella", "Zev", "Zinnia", "Zen", "Zendaya", "Zavier", "Zia", "Zach", "Zuri"},
}

// NameGenerator holds the worker pool for name generation
type NameGenerator struct {
	pool              *workerpool.WorkerPool
	nameCacheMutex    sync.RWMutex
	nameCache         map[string][]string // Cache for previously generated names
	nameGeneratorSeed int64
}

// NewNameGenerator creates a new name generator with a worker pool
func NewNameGenerator(numWorkers int) *NameGenerator {
	// Create a new worker pool
	pool := workerpool.New(numWorkers)
	
	// Create a new name generator
	generator := &NameGenerator{
		pool:              pool,
		nameCache:         make(map[string][]string),
		nameGeneratorSeed: time.Now().UnixNano(),
	}
	
	return generator
}

// DefaultGenerator is the default global name generator instance
var (
	DefaultGenerator     *NameGenerator
	defaultGeneratorOnce sync.Once
)

// GetDefaultGenerator returns the default name generator instance
func GetDefaultGenerator() *NameGenerator {
	defaultGeneratorOnce.Do(func() {
		// Use number of CPU cores as the number of workers
		numWorkers := 4
		DefaultGenerator = NewNameGenerator(numWorkers)
	})
	
	return DefaultGenerator
}

// getCacheKey returns a cache key for the given letter and count
func getCacheKey(letter string, count int) string {
	return letter + ":" + string(rune(count))
}

// GenerateNames generates a list of random names starting with the specified letter
// This is now just a wrapper around the default generator
func GenerateNames(letter string, count int) []string {
	return GetDefaultGenerator().Generate(letter, count)
}

// GenerateNamesWithContext generates names with a context for cancellation
func GenerateNamesWithContext(ctx context.Context, letter string, count int) []string {
	return GetDefaultGenerator().GenerateWithContext(ctx, letter, count)
}

// Generate generates a list of random names starting with the specified letter
func (g *NameGenerator) Generate(letter string, count int) []string {
	// Create a default context with a reasonable timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	return g.GenerateWithContext(ctx, letter, count)
}

// GenerateWithContext generates a list of random names with a context for cancellation
func (g *NameGenerator) GenerateWithContext(ctx context.Context, letter string, count int) []string {
	// If count is zero or negative, return empty slice
	if count <= 0 {
		return []string{}
	}
	
	// If no letter is specified, choose one randomly
	if letter == "" {
		letters := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}
		letter = letters[rand.Intn(len(letters))]
	} else {
		// Convert letter to uppercase
		letter = strings.ToUpper(string(letter[0]))
	}
	
	// Get the list of names for the specified letter
	namesList, ok := NamesByLetter[letter]
	if !ok || len(namesList) == 0 {
		// If no names exist for this letter, return an empty slice
		return []string{}
	}
	
	// If count is greater than the available names, limit it
	if count > len(namesList) {
		count = len(namesList)
	}
	
	// Check if the names are already in the cache
	cacheKey := getCacheKey(letter, count)
	g.nameCacheMutex.RLock()
	cachedNames, found := g.nameCache[cacheKey]
	g.nameCacheMutex.RUnlock()
	
	if found && len(cachedNames) >= count {
		// Return a copy of the cached names to avoid data races
		result := make([]string, count)
		copy(result, cachedNames[:count])
		return result
	}
	
	// Generate random names in parallel using the worker pool
	names := make([]string, count)
	tasks := make([]workerpool.Task, count)
	
	// Create a task for each name generation
	for i := 0; i < count; i++ {
		index := i // Capture the index in the closure
		tasks[i] = func() interface{} {
			// Create a source of randomness that's isolated to this task
			taskRand := rand.New(rand.NewSource(time.Now().UnixNano() + int64(index)))
			randomIndex := taskRand.Intn(len(namesList))
			return namesList[randomIndex]
		}
	}
	
	// Submit tasks in batch and get results
	resultCh := g.pool.SubmitBatch(tasks)
	
	// Process results as they come in
	i := 0
	for result := range resultCh {
		if i >= count {
			break
		}
		
		// Check if the context has been canceled
		select {
		case <-ctx.Done():
			// Context canceled, return what we have so far
			return names[:i]
		default:
			// Continue processing
		}
		
		// Get the name from the result
		name, ok := result.Value.(string)
		if ok {
			names[i] = name
			i++
		}
	}
	
	// Update the cache with the generated names
	g.nameCacheMutex.Lock()
	g.nameCache[cacheKey] = make([]string, len(names))
	copy(g.nameCache[cacheKey], names)
	g.nameCacheMutex.Unlock()
	
	return names
}

// Shutdown gracefully shuts down the name generator's worker pool
func (g *NameGenerator) Shutdown() {
	g.pool.Shutdown()
}

// ShutdownNow immediately shuts down the name generator's worker pool
func (g *NameGenerator) ShutdownNow() {
	g.pool.ShutdownNow()
}
