package service

import "strings"

// Define activity types based on Strava sport_type
const (
	// Basic types
	Run            = "Run"
	Ride           = "Ride"
	Swim           = "Swim"
	Walk           = "Walk"
	Workout        = "Workout"
	WeightTraining = "WeightTraining"
	Yoga           = "Yoga"

	// Additional types
	Hike             = "Hike"
	TrailRun         = "TrailRun"
	VirtualRide      = "VirtualRide"
	VirtualRun       = "VirtualRun"
	Elliptical       = "Elliptical"
	StairStepper     = "StairStepper"
	Crossfit         = "Crossfit"
	Pilates          = "Pilates"
	Skateboard       = "Skateboard"
	Surf             = "Surf"
	Soccer           = "Soccer"
	Squash           = "Squash"
	MountainBikeRide = "MountainBikeRide"
	Canoeing         = "Canoeing"

	Default = "Default"
)

// Update the activity type detection
func getActivityType(activityName string, sportType string) string {
	// First try to match by sport_type if available
	switch sportType {
	case Run, Ride, Swim, Walk, Workout, WeightTraining, Yoga,
		Hike, TrailRun, VirtualRide, VirtualRun, Elliptical,
		StairStepper, Crossfit, Pilates, Skateboard, Surf,
		Soccer, Squash, MountainBikeRide, Canoeing:
		return sportType
	}

	// Fallback to name-based detection for backward compatibility
	switch {
	case strings.Contains(activityName, "Run"):
		return Run
	case strings.Contains(activityName, "Ride"):
		return Ride
	case strings.Contains(activityName, "Swim"):
		return Swim
	case strings.Contains(activityName, "Walk"):
		return Walk
	case strings.Contains(activityName, "Weight Training"):
		return WeightTraining
	case strings.Contains(activityName, "Yoga"):
		return Yoga
	default:
		return Default
	}
}

var activityJokes = map[string][]string{
	Run: {
		"🏃‍♀️ Eu vou correr (pra te ver) - Charlie Brown Jr da academia",
		"🏃‍♂️ Se correr o bicho pega, se ficar sem cardio o shape pega",
		"🏃‍♀️ Naruto é meu personal trainer, dattebayo!",
		"🏃‍♂️ Sonic depois do açaí com pão de queijo",
		"🏃‍♂️ Acelerou e foi embora - Ana Carolina no longão",
		"🏃‍♀️ Vou deixar a vida me levar (correndo) - Zeca Pagodinho no treino",
		"🏃‍♂️ Corra, que a vida é passageira - Cássia Eller no fartlek",
		"🏃‍♀️ Encontrei a corrida no meio do caminho - Lulu Santos no aquecimento",
		"🏃‍♂️ Vou correr na avenida Brasil - Tim Maia no treino de 5k",
		"🏃‍♀️ Aonde quer que eu vá (eu corro) - Paralamas no treino intervalado",
		"🏃‍♂️ Foi só correr pra te encontrar - Maria Rita na trilha",
		"🏃‍♀️ Anunciação (de um novo pace) - Alceu Valença no longão",
		"🏃‍♂️ Vem correr comigo - Los Hermanos no cooper",
		"🏃‍♀️ Amanhã (vai ser melhor correr) - Guilherme Arantes no treino regenerativo",
		"🏃‍♂️ Só quero correr mais um pouquinho - Djavan no sprint final",
		"🏃‍♀️ Velocidade do meu coração (correndo) - Roupa Nova no treino de pista",
		"🏃‍♂️ À primeira corrida - Elis Regina no treino inicial",
		"🏃‍♀️ Se você correr bem devagar, ainda assim vai chegar - Engenheiros do Hawaii na planilha",
		"🏃‍♂️ Corre atrás do sol - Oswaldo Montenegro no treino de final de tarde",
		"🏃‍♀️ Tempo perdido (correndo no ritmo errado) - Legião Urbana no fartlek",
		"🏃‍♂️ Minha alma (não corre sozinha) - O Rappa na prova de revezamento",
		"🏃‍♀️ Mais perto que nunca (do pódio) - Biquini Cavadão na corrida",
		"🏃‍♂️ É preciso correr pra viver - Vinícius de Moraes no longão",
		"🏃‍♀️ Quem sabe correr faz a hora - Geraldo Vandré no pace estável",
		"🏃‍♂️ Toda forma de correr vale a pena - Titãs no treino alternativo",
		"🏃‍♀️ Nem sempre ganho, mas corro - Marisa Monte no treino mental",
		"🏃‍♂️ Tô correndo atrás dos meus sonhos - Seu Jorge na trilha",
		"🏃‍♀️ Vamos correr (por onde for) - Ana Carolina no treino leve",
		"🏃‍♂️ Hoje corri pra me sentir vivo - Milton Nascimento no nascer do sol",
		"🏃‍♀️ Na corrida dos meus sonhos - Vanessa da Mata no treino noturno",
		"🏃‍♂️ Cada pace um caminho - Os Paralamas no treino experimental",
		"🏃‍♀️ Devagar e sempre, vou correndo - Almir Sater no regenerativo",
		"🏃‍♂️ Enquanto houver corrida, há caminho - Zé Ramalho na ultramaratona",
		"🏃‍♀️ Correndo no céu azul - Djavan no treino de domingo",
		"🏃‍♂️ A linha de chegada é só o começo - Barão Vermelho na maratona",
		"🏃‍♀️ Pro dia nascer feliz (correndo) - Cazuza no treino matinal",
		"🏃‍♂️ Vai correr e me deixa aqui - Capital Inicial no descanso ativo",
		"🏃‍♀️ Enquanto eu corro, tudo se encaixa - Maria Bethânia na trilha",
		"🏃‍♂️ Te vejo na largada, mas vou te passar - Chico Science no tiro",
		"🏃‍♀️ O que é que a corrida tem? - Caetano Veloso no longão",
		"🏃‍♂️ Me deixa correr (em paz) - Adriana Calcanhotto no treino solitário",
		"🏃‍♀️ Vida que segue correndo - Milton Nascimento na planilha",
		"🏃‍♂️ Corra atrás de mim se puder - Raul Seixas no fartlek",
		"🏃‍♀️ Deixa a vida me levar, correndo - Zeca Pagodinho na trilha leve",
		"🏃‍♂️ Não pare na pista - Erasmo Carlos no progressivo",
		"🏃‍♀️ Eu só corro porque amo - Djavan no regenerativo",
		"🏃‍♂️ Por onde correr (não importa) - Legião Urbana no treino de resistência",
		"🏃‍♀️ Me espera na chegada - Sandy e Junior na corrida de casais",
		"🏃‍♂️ Só quero correr mais um quilômetro - Tim Maia no sprint final",
		"🏃‍♀️ Você me faz correr (do meu melhor) - Frejat no treino forte",
		"🏃‍♂️ O ritmo da minha corrida - Ana Carolina no treino técnico",
		"🏃‍♀️ O sol na trilha da corrida - Gilberto Gil no treino matutino",
		"🏃‍♂️ Corri tanto pra chegar - Skank na meia maratona",
	},
	Ride: {
		"🚴‍♂️ Pedalando na garupa da saudade",
		"🚴‍♀️ Quem não tem carro, pedala como o ET",
		"🚴‍♂️ Domingo eu vou pedalar de Bicicleta (Mashup Turma do Pagode)",
		"🚴‍♀️ Pedala, pedala, pedala, meu bem (Tim Maia Feelings)",
		"🚴‍♀️ Rainha da South (bike) versão Glória Groove",
		"🚴‍♂️ Graças a Deus sou ciclista (Jorge Ben Feelings)",
		"🚴‍♀️ Meu pedal é tipo Balão Mágico: Sempre pra frente",
		"🚴‍♀️ Por toda a minha vida, eu vou pedalar - Elis Regina no longão",
		"🚴‍♂️ Vai pedalando sem medo de ser feliz - Lulu Santos na descida",
		"🚴‍♀️ Cada curva dessa estrada é minha - Djavan no treino solo",
		"🚴‍♂️ Eu só quero uma bike no fim da tarde - Almir Sater na trilha",
		"🚴‍♀️ Pedalando no céu azul - Roupa Nova no longão matinal",
		"🚴‍♂️ Liberdade pra dentro da bike - Natiruts no treino regenerativo",
		"🚴‍♀️ Tudo começou num pedal pela cidade - Adriana Calcanhotto na ciclovia",
		"🚴‍♂️ Eu não vou parar, vou pedalar mais longe - Barão Vermelho no treino de resistência",
		"🚴‍♀️ Todo dia ela pedala cedo - Jorge Ben Jor na estrada",
		"🚴‍♂️ Não diga que a bike não faz sentido - Engenheiros do Hawaii no revezamento",
		"🚴‍♀️ O que é que a bike tem que tanto encanta? - Caetano Veloso no passeio leve",
		"🚴‍♂️ Vou pedalar até o sol raiar - Vanessa da Mata no treino noturno",
		"🚴‍♀️ Quero ver o sol nascer do guidão da bike - Milton Nascimento na subida",
		"🚴‍♂️ A vida é feita pra pedalar - Lenine na trilha",
		"🚴‍♀️ Mais uma volta e eu chego lá - Sandy e Junior na subida final",
		"🚴‍♂️ O ritmo das pedaladas me faz sonhar - Ana Carolina no treino técnico",
		"🚴‍♀️ Nada vai me fazer parar de pedalar - Titãs no progressivo",
		"🚴‍♂️ Por onde for, leve sua bike - Cidade Negra na viagem",
		"🚴‍♀️ É devagar, é devagarinho, na bike - Martinho da Vila no treino leve",
		"🚴‍♂️ Sempre há um novo horizonte pra pedalar - Zé Ramalho na ultramaratona",
		"🚴‍♀️ Hoje eu só quero pedalar em paz - Roberto Carlos na trilha",
		"🚴‍♂️ Lá vou eu, na estrada da bike - Ivete Sangalo no passeio",
		"🚴‍♀️ Vai, pedala e não olha pra trás - Capital Inicial no treino de força",
		"🚴‍♂️ De bicicleta eu vou, por onde o vento levar - Arnaldo Antunes no pedal matinal",
		"🚴‍♀️ Viver e pedalar, tudo junto - Marisa Monte no passeio em dupla",
		"🚴‍♂️ Vamos pedalar e fazer a hora - Geraldo Vandré no grupetto",
		"🚴‍♀️ No balanço da bike, eu encontrei paz - Os Paralamas no regenerativo",
		"🚴‍♂️ Quando o sol nascer, lá estarei de bike - Gilberto Gil no treino",
		"🚴‍♀️ Com a bike, a vida é mais bonita - Toquinho no passeio da manhã",
		"🚴‍♂️ Pedalar é preciso, viver também - Caetano Veloso na jornada",
		"🚴‍♀️ Me deixa pedalar, vai - Adriana Calcanhotto na subida",
		"🚴‍♂️ Só ando de bike por aí - Jorge Ben Jor na ciclovia",
		"🚴‍♀️ Leve a vida como uma trilha - Legião Urbana no pedal na natureza",
		"🚴‍♂️ O amor pela bike ninguém tira - Cazuza no treino livre",
		"🚴‍♀️ Subindo a serra, pedalando e cantando - Alceu Valença na subida desafiadora",
		"🚴‍♂️ Longe, mas sempre de bike - Skank no passeio em grupo",
		"🚴‍♀️ Bike no ritmo da estrada - Djavan no treino longo",
		"🚴‍♂️ Vou pedalando e deixando o tempo passar - Tiago Iorc no treino noturno",
		"🚴‍♀️ Quero ver o mundo do guidão - Nando Reis no pedal explorador",
		"🚴‍♂️ De bike, a vista é mais bonita - Vanessa da Mata na viagem longa",
		"🚴‍♀️ A liberdade mora no pedal - Oswaldo Montenegro na trilha solo",
		"🚴‍♂️ Pedalei até onde o horizonte me chamou - Milton Nascimento na ultra",
		"🚴‍♀️ Por entre curvas e subidas, vou de bike - Gilberto Gil no treino",
		"🚴‍♂️ A estrada é longa, mas a bike me leva - Frejat no progressivo",
		"🚴‍♀️ Só peço um dia de pedal e céu azul - Roupa Nova no passeio leve",
		"🚴‍♂️ Na roda da bike, encontro paz - Djavan na trilha",
		"🚴‍♀️ De bike, o amor é mais simples - Ana Carolina no passeio matutino",
		"🚴‍♂️ A vida me chama pra pedalar - Seu Jorge na ciclovia",
		"🚴‍♀️ Um pedal pra esquecer o mundo - Maria Rita na manhã tranquila",
		"🚴‍♂️ O vento no rosto, só de bike - Chico César no treino regenerativo",
	},
	Swim: {
		"🏊‍♀️ Vai nadar pra ver se aprende - Sandy no aquecimento",
		"🏊‍♂️ Água no meu caminho - Djavan no treino livre",
		"🏊‍♀️ Desliza, desliza - Ivete Sangalo no estilo crawl",
		"🏊‍♂️ Fui nadando contra a maré, sem saber voltar - Lenine no treino de resistência",
		"🏊‍♀️ Na onda da natação, eu vou - Netinho no recreativo",
		"🏊‍♂️ Águas de março nas braçadas - Tom Jobim na piscina",
		"🏊‍♀️ Vai na marola, vai - Armandinho no nado costas",
		"🏊‍♂️ Atravessar o oceano no peito - Zeca Baleiro no treino de longa distância",
		"🏊‍♀️ Eu sei nadar na beira do mar - Caetano Veloso no regenerativo",
		"🏊‍♂️ Água de beber (da piscina, é claro) - Vinícius de Moraes no descanso",
		"🏊‍♀️ Braçada e perna, tudo tem seu momento - Gilberto Gil no estilo medley",
		"🏊‍♂️ E eu nadava, nadava contra o tempo - Skank no treino forte",
		"🏊‍♀️ Vou pra água lavar a alma - Maria Gadú no treino inicial",
		"🏊‍♂️ Até o fundo do mar, eu chego - Chico Buarque no mergulho",
		"🏊‍♀️ A piscina me chama e eu vou - Roberto Carlos no recreativo",
		"🏊‍♂️ Me molha de amor e mar - Jota Quest no treino",
		"🏊‍♀️ Flutuar é mais fácil do que parece - Marisa Monte no aquecimento",
		"🏊‍♂️ Nado crawl, peito e borboleta - Alceu Valença na aula",
		"🏊‍♀️ Vai buscar o ar onde for preciso - Engenheiros do Hawaii no fôlego",
		"🏊‍♂️ Tudo que é água é minha casa - Os Paralamas na travessia",
		"🏊‍♂️ Vai precisar de um barco maior... - Jaws no aquecimento",
		"🏊‍♀️ Me siga se puder! - Forrest Gump no estilo borboleta",
		"🏊‍♂️ O mar está chamando! - Moana na travessia",
		"🏊‍♀️ Um verdadeiro mago nunca nada apressado - Gandalf no aquecimento",
		"🏊‍♂️ Com grandes poderes vêm grandes braçadas - Homem-Aranha na piscina",
		"🏊‍♀️ Nunca deixe ir, Jack... nunca! - Rose no nado sincronizado",
		"🏊‍♂️ Só eu nado no escuro? - Batman na aula noturna",
		"🏊‍♀️ Vou fazer uma oferta que você não pode recusar: nadar! - O Poderoso Chefão no convite",
		"🏊‍♂️ Nadar ou não nadar, eis a questão! - Hamlet na piscina",
		"🏊‍♀️ Continue nadando, continue nadando! - Dory no treino de constância",
		"🏊‍♂️ Eu vejo água por todo lado - O Sexto Sentido na maratona aquática",
		"🏊‍♀️ Meu precioso... minha piscina! - Gollum no aquecimento",
		"🏊‍♂️ É só um pequeno passo até a borda! - Neil Armstrong na aula de iniciante",
		"🏊‍♀️ Eu voltarei... pra nadar mais rápido! - Exterminador do Futuro no sprint",
		"🏊‍♂️ A força estará com você, sempre! - Yoda no dia de treino pesado",
		"🏊‍♀️ A água é clara como a luz do sol! - Avatar na aula de mergulho",
		"🏊‍♂️ No fundo do mar, o que nos resta é nadar! - Titanic no treino livre",
	},
	WeightTraining: {
		"🏋️ Supino supimpa supera superação",
		"💪 Suando mais que político em CPI",
		"💪 Malhando mais que o Bambam no auge",
		"🏋️‍♀️ Levantando mais peso que as minhas escolhas",
		"🏋️‍♂️ O fardo pesa, mas não posso parar - Senhor dos Anéis no agachamento",
		"🏋️‍♀️ Eu vejo pesos por todo lado - O Sexto Sentido na academia",
		"🏋️‍♂️ Levanta, sacode a poeira e dá mais uma série - Beth Carvalho no leg press",
		"🏋️‍♀️ Que a força esteja com você! - Star Wars no supino",
		"🏋️‍♂️ Só eu e o ferro, no balanço da vida - Djavan no levantamento terra",
		"🏋️‍♀️ Com grandes pesos vêm grandes resultados - Homem-Aranha na rosca direta",
		"🏋️‍♂️ Eu sou o rei do agachamento! - Titanic no rack",
		"🏋️‍♀️ Vem malhar, meu bem querer - Djavan na puxada alta",
		"🏋️‍♂️ This is Sparta! - 300 no leg day",
		"🏋️‍♀️ Vai, levanta esse peso devagarinho - Martinho da Vila no deadlift",
		"🏋️‍♂️ Pesado é o que carrego no coração (e no supino) - Lulu Santos na academia",
		"🏋️‍♀️ Eu só quero malhar mais um pouquinho - Tim Maia no drop set",
		"🏋️‍♂️ Nunca desista da barra! - Capitão América no treino",
		"🏋️‍♀️ Ninguém solta o halter de ninguém - Marisa Monte no treino funcional",
		"🏋️‍♂️ O agachamento é meu lar, meu refúgio - Legião Urbana no treino",
		"🏋️‍♀️ I'll be back... pro próximo treino - Exterminador do Futuro no HIIT",
		"🏋️‍♂️ Força e fé pra mais uma repetição - Zeca Pagodinho no supino reto",
		"🏋️‍♀️ Aumenta o peso, mas com carinho - Ivete Sangalo no leg press",
		"🏋️‍♂️ A barra é pesada, mas eu sou forte! - Superman na puxada baixa",
		"🏋️‍♀️ Eu sou inevitável... no treino de ombros - Thanos na rosca martelo",
		"🏋️‍♂️ Vai malhando até onde puder - Sandy & Junior no treino",
		"🏋️‍♀️ Eu vou levantar, não importa o peso - Vinícius de Moraes no terra",
		"🏋️‍♂️ É no supino que a mágica acontece - Harry Potter no treino",
		"🏋️‍♀️ Um peso de cada vez - Forrest Gump na academia",
		"🏋️‍♂️ Eu treino, logo existo - Hamlet na sala de pesos",
		"🏋️‍♀️ Braço de ferro e alma de aço - Roberto Carlos no bíceps curl",
		"🏋️‍♂️ Deixa a força me levar - Zeca Pagodinho no treino de peito",
		"🏋️‍♀️ Os pesos não mentem jamais - Drauzio Varella no agachamento",
		"🏋️‍♂️ A subida é dura, mas o resultado é doce - Senhor dos Anéis no leg day",
		"🏋️‍♀️ Braços fortes, coração leve - Nando Reis no treino",
		"🏋️‍♂️ Por que tão pesado? - Coringa na academia",
		"🏋️‍♀️ Tudo que é duro vale a pena - Engenheiros do Hawaii no treino",
		"🏋️‍♂️ Vai buscar o shape que é seu - Os Paralamas no drop set",
		"🏋️‍♀️ Subindo peso, descendo ego - Elis Regina no rack",
		"🏋️‍♂️ Quanto mais peso, mais amor - Ana Carolina na sala de musculação",
		"🏋️‍♀️ Só o agachamento salva! - Matrix no leg day",
		"🏋️‍♂️ Deixe o ferro me guiar - Marisa Monte no treino de bíceps",
		"🏋️‍♀️ Um halter de cada vez, e tudo se ajeita - Arnaldo Antunes no treino funcional",
		"🏋️‍♂️ O importante é começar leve e terminar forte - Ivete Sangalo no supino",
		"🏋️‍♀️ Esse peso vai cair... só no chão! - Dom Casmurro no deadlift",
		"🏋️‍♂️ Puxando peso, construindo sonhos - Milton Nascimento no pulley",
		"🏋️‍♀️ A barra nunca me abandona - Nando Reis na rosca direta",
		"🏋️‍♂️ Eu levanto o que ninguém vê - Toquinho na sala de musculação",
		"🏋️‍♀️ No agachamento, encontrei meu destino - Gilberto Gil no treino",
		"🏋️‍♂️ Malhando no ritmo da vida - Paralamas do Sucesso no bíceps curl",
		"🏋️‍♀️ Eu não caio, só agacho! - Dory no treino de perna",
		"🏋️‍♂️ Só treino pra ter paz - Seu Jorge no crossover",
		"🏋️‍♀️ Pesado é pouco pra quem tem foco - Beyoncé na academia",
		"🏋️‍♂️ No ferro, eu sou imbatível - Rocky no treino",
		"🏋️‍♀️ Vai devagar e sempre no levantamento - Caetano Veloso no treino",
	},
	Yoga: {
		"🧘‍♂️ Namastê mais zen que nunca",
		"🧘‍♀️ Minha vida é um chakra que não fecha",
		"🧘‍♂️ Ohm meu Deus, que alongamento",
		"🧘‍♀️ Medita, respira e não pira",
		"🧘‍♂️ Yoga é tipo RPG: cada dia subo um level",
	},
	Walk: {
		"🚶‍♂️ Caminhando e cantarolando pela vida",
		"🚶‍♀️ Dando uma de Faustão: Olha o passeio",
		"🚶‍♂️ Andando mais que o Silvio Santos no palco",
		"🚶‍♀️ Minha vida é uma novela das 7: muita caminhada",
		"🚶‍♂️ Passos contados, stories postados",
		"🚶‍♀️ Caminhando mais que o Zeca Pagodinho no bar",
		"🚶‍♂️ Andando tipo Roberto Carlos: Além do Horizonte",
		"🚶‍♀️ Passos mais certeiros que dancinha do TikTok",
		"🚶‍♂️ Caminhando na paz do Senhor do Bom Fim",
		"🚶‍♀️ Caminhando na batida do Molejo",
		"🚶‍♂️ Andando mais que fiscal do Detran",
		"🚶‍♂️ Caminhando na velocidade do Internet Explorer",
		"🚶‍♀️ Andando mais que Pokémon sem Pokébola",
		"🚶‍♂️ Passeando tipo Caetano: Sozinho",
		"🚶‍♂️ Andando mais que vendedor de pamonha",
	},
	VirtualRide: {
		"🚴‍♂️ Pedalando no metaverso da magreza",
		"🚴‍♂️ Pedalando em Nárnia virtual",
		"🚴‍♀️ Meu avatar pedala melhor que eu",
	},
	Crossfit: {
		"💪 WOD: Workout Of Destruição",
		"🏋️‍♂️ Crossfit é tipo Carrefour: tem de tudo",
		"💪 Box mais intenso que novela da Glória Perez",
		"🏋️‍♀️ Fazendo mais repetição que funk do verão",
	},
	MountainBikeRide: {
		"🚵‍♂️ Pedalando mais que o Louro José fugindo da Ana Maria",
		"🚵‍♀️ Trilha sonora da minha vida tem pedal e poeira",
		"🚵‍♂️ Aventura com dois aros e muita história",
		"🚵‍♀️ Pedalando onde o vento faz a curva",
	},
	Default: {
		"💪 Suando mais que político em CPI",
		"🎯 Malhação: Vibe Fitness",
		"💫 Treinando mais que participante do BBB",
		"🌟 Academia é meu Big Brother particular",
	},
}
