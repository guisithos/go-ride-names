package service

// Define activity types
const (
	Run     = "Run"
	Ride    = "Ride"
	Swim    = "Swim"
	Walk    = "Walk"
	Workout = "WeightTraining"
	Yoga    = "Yoga"
	Default = "Default"
)

// Organize jokes by activity type
var activityJokes = map[string][]string{
	Run: {
		"🏃‍♂️ Running late is my cardio",
		"🏃‍♀️ Running from my problems (and catching up)",
		"🏃‍♂️ Chasing my dreams (they're pretty fast)",
		"🏃‍♀️ Professional pizza burner",
		"🏃‍♂️ Running on caffeine and determination",
		"🏃‍♀️ Running like there's cake at the finish line",
		"🏃‍♂️ Running from adult responsibilities",
		"🏃‍♀️ Running on empty but still going",
		"🏃‍♂️ Running from my comfort zone",
		"🏃‍♀️ Running late is still running",
		"🏃‍♂️ These legs were made for running",
	},
	Ride: {
		"🚴‍♂️ Bike to the future",
		"🚴‍♀️ Wheel-y good workout",
		"🚴‍♂️ Pedal to the metal (but it's a bicycle)",
		"🚴‍♂️ Two tired to stop (get it?)",
		"🚴‍♀️ Spoke too soon about being fit",
		"🚴‍♂️ Chain reaction to fitness",
		"🚴‍♀️ Ride and shine",
		"🚴‍♂️ Wheel power",
	},
	Swim: {
		"🏊‍♂️ Just keep swimming, just keep drowning my sorrows",
		"🏊‍♀️ Pool party of one",
		"🏊‍♂️ Swimming in endorphins",
		"🏊‍♀️ Becoming a mermaid, one lap at a time",
		"🏊‍♂️ Just add water and motivation",
		"🏊‍♂️ Swimming in success (and chlorine)",
		"🏊‍♀️ Making waves and progress",
		"🏊‍♂️ Pool's out for summer",
		"🏊‍♀️ Just keep splashing",
	},
	Walk: {
		"🚶‍♂️ Walking because running apps crash",
		"🚶‍♀️ Walking off the pizza from last night",
		"🚶‍♂️ Walking because running was too mainstream",
		"🚶‍♀️ Step by step to greatness",
		"🚶‍♂️ Small steps, big dreams",
	},
	Workout: {
		"🏋️ These weights aren't going to lift themselves... unfortunately",
		"💪 I'm not sweating, I'm leaking awesomeness",
		"🏋️‍♀️ Lifting spirits and heavy things",
		"💪 Making my muscles cry",
		"🏋️ Getting stronger than my excuses",
		"💪 Turning fat into abs and tears",
		"🏋️‍♂️ Lifting weights and spirits",
		"💪 Making muscles, making memories",
		"🏋️‍♀️ Beast mode with a side of sass",
		"💪 Flex appeal in progress",
		"🏋️ Weight for it... getting stronger",
		"💪 Muscle hustle",
		"🏋️‍♂️ Lifting my way to legendary",
		"💪 No pain, no gain, no kidding",
		"🏋️ Weight a minute, I'm not done",
		"💪 Getting fit-ish",
		"🏋️‍♀️ Dumbbells and smart moves",
	},
	Yoga: {
		"🧘‍♀️ Namaste in bed, but I'm here",
		"🧘‍♂️ Getting my zen on (and trying not to fall)",
		"🧘‍♀️ Pretzel in progress",
		"🧘‍♂️ Stretching the limits of possibility",
		"🧘‍♀️ Finding my inner peace (it's hiding pretty well)",
		"🧘‍♂️ Yoga now, wine later",
		"🧘‍♀️ Becoming one with my mat",
		"🧘‍♂️ Bending so I don't break",
		"🧘‍♀️ Channeling my inner guru",
		"🧘‍♂️ Warrior pose, peaceful mind",
		"🧘‍♀️ Stretching the boundaries of reality",
		"🧘‍♂️ Yoga: because therapy is expensive",
	},
	Default: {
		"💪 Making progress, one workout at a time",
		"🎯 Another day, another goal crushed",
		"💫 Living my best active life",
		"🌟 Plot twist: getting stronger",
		"💪 Level up in progress",
	},
}
