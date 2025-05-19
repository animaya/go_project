package generator

import (
	"math/rand"
	"strings"
	"sync"
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

// GenerateNames generates a list of random names starting with the specified letter
func GenerateNames(letter string, count int) []string {
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

	// Generate random names in parallel
	names := make([]string, count)
	var wg sync.WaitGroup
	wg.Add(count)
	
	for i := 0; i < count; i++ {
		go func(index int) {
			defer wg.Done()
			// Randomly select a name from the list
			randomIndex := rand.Intn(len(namesList))
			names[index] = namesList[randomIndex]
		}(i)
	}
	
	wg.Wait()
	
	return names
}
