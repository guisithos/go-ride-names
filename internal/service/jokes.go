package service

import "strings"

// Define activity types based on Strava sport_type
const (
	// Basic types
	Run            = "Run"
	Ride           = "Ride"
	Swim           = "Swim"
	Walk           = "Walk"
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
	Workout          = "Workout"

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
		"🏃‍♂️ 'Minha corrida é como meu código: cheia de loops infinitos e erros inesperados.'",
		"👟 'Tentei correr como no Counter-Strike, mas esqueci que na vida real não existe bunny hop.'",
		"🏃‍♀️ 'Meu pace é tão lento que o lag do servidor pensou que eu tinha desconectado.'",
		"🕹️ 'Correr é como compilar um programa: demora, dá erro, mas eventualmente funciona.'",
		"🎮 'Se houvesse um modo fácil na vida real, corrida seria um quicksave antes de cada ladeira.'",
		"⚔️ 'Eu corro com a mesma estratégia de um bárbaro de D&D: tudo na força, zero na destreza.'",
		"🏃‍♂️ 'A diferença entre correr e programar? No código, você pode debugar; na corrida, só sofre.'",
		"🖱️ 'Correndo me sinto no Dota: muita ação, mas no final meu time (meu corpo) me deixa na mão.'",
		"🎲 'Teste de resistência na corrida? Rolei 1 crítico e tropecei no próprio cadarço.'",
		"🏃‍♀️ 'Corrida longa é como um RPG de turno: decisões lentas e dor a cada movimento.'",
		"🔧 'Preciso de um script em Python para automatizar minhas pernas. Esse loop manual está ineficiente.'",
		"🌌 'Correr é como jogar Skyrim: você começa empolgado, mas logo quer fast travel até o final.'",
		"🎮 'Eu corro como um bot do CS: reto para a parede, sem desviar dos obstáculos.'",
		"🧙‍♂️ 'Se fosse um mago, eu usaria teleport. Mas não, sou só um humano com pouca estamina.'",
		"🏃‍♂️ 'Correr é o debug da vida: a cada erro você fica mais perto de uma solução (ou da desistência).'",
		"🛡️ 'Na corrida, sou um tanque de RPG: movo devagar, mas aguento bastante dano emocional.'",
		"🎲 'Rolei iniciativa para correr, mas o mestre deu uma ladeira de desvantagem.'",
		"🖱️ 'Meu pace é tão ruim que no LoL seria considerado feeding.'",
		"🎮 'Correr na chuva me faz sentir no GTA: escorregando e batendo em tudo sem controle.'",
		"🏃‍♀️ 'Se corrida fosse multiplayer, eu seria o cara carregado na partida.'",
		"🕹️ 'Correr é como grindar XP: chato, repetitivo, mas eventualmente você level up.'",
		"🔍 'A diferença entre programação e corrida? Na primeira eu só travo, na segunda eu travo e caio.'",
		"🏃‍♂️ 'Depois de uma corrida, minha stack overflow é muscular.'",
		"⚔️ 'Meu DM disse que correr era bom para stamina. Ele esqueceu de me avisar sobre a dor eterna.'",
		"🎮 'Correr na esteira é como o loading screen: a sensação de ir a lugar nenhum.'",
		"🧙‍♂️ 'Tentei correr como um rogue. Esqueci que não tenho stealth nem agilidade.'",
		"🏃‍♀️ 'Na corrida, meu pace é tão lento que pareço um NPC de fetch quest.'",
		"🎲 'Se corrida fosse uma rolagem de dados, minha constituição seria -2.'",
		"🕹️ 'Correr é como no Dota: você tenta fugir, mas sempre tem uma torre (colina) para te acabar.'",
		"⚔️ 'Correr de manhã é uma side quest: muita dificuldade por pouca recompensa.'",
		"🏃‍♂️ 'Me inscrevi para uma corrida. Parecia um evento bônus, mas virou um boss fight.'",
		"🎮 'Meu pace é tipo conexão dial-up: lento, instável e com muitas quedas.'",
		"🖱️ 'Corrida é um bug no meu sistema: pernas não conectam com motivação.'",
		"🏃‍♀️ 'Tentei correr full stack, mas fiquei preso no front-end: as pernas.'",
		"🎮 'Se correr é um jogo, a minha dificuldade está setada em pesadelo.'",
		"🌌 'Corrida me lembra No Man's Sky: interminável e cheia de frustrações.'",
		"⚔️ 'Se fosse uma quest, correr seria hardcore mode: 1 erro e você sente por uma semana.'",
		"🏃‍♂️ 'Meu pace é como um servidor sem cache: lento e constante.'",
		"🖱️ 'Corro como um mago com lag: muito preparo, pouco movimento.'",
		"🎮 'Corrida é como farmar em MMORPG: lenta e dolorosa, mas alguém diz que vale a pena.'",
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
		"🏊‍♂️ 'Continue a nadar, continue a nadar...' - Procurando Nemo, meu mantra na série de crawl.",
		"💪 'Sou o Aquaman da piscina: poderoso até que alguém ligue o filtro.'",
		"🏋️‍♀️ 'Nadar é como Titanic: começa tranquilo, mas no final, você afunda.",
		"🦾 'Eu sou a Pequena Sereia da piscina, só que troquei o canto pelas braçadas.",
		"🏊‍♂️ *Treinar natação é como Tubarão: 'A dor está sempre à espreita.'",
		"💥 *'Nadar borboleta é tipo Moana: 'O mar te chama, mas a correnteza te segura.'",
		"🏃‍♂️ *A piscina é meu Náufrago: 'Wilson é minha touca, sempre junto.'",
		"🦵 *Nado costas é meu 'Piratas do Caribe': tentando manter o tesouro (fôlego).",
		"💪 'É um pássaro? Um avião? Não, sou eu tentando nadar crawl sem beber água.'",
		"🏊‍♀️ *Cada virada na borda é como Life of Pi: 'A luta pela sobrevivência começa.'",
		"💦 *'Tentei treinar peito, mas foi Titanic: 'Afundei na segunda piscina.'",
		"🏋️‍♂️ *'A piscina é meu Sharknado: 'Se não é o treino, é o medo de engolir água.'",
		"💪 *Nadar borboleta é tipo Avatar: 'Você tenta ser uma lenda, mas só vê azul.'",
		"🏊‍♂️ 'Quem vive na piscina de alguém? Eu no dia de treino!' - versão Bob Esponja.",
		"🦾 *Natação é meu Free Willy: 'Sempre tentando saltar as barreiras (do cansaço).'",
		"💦 *Cada 100m na piscina é como Aquaman: 'Entre dois mundos e sem ar em nenhum.'",
		"🏊‍♀️ *Treino de crawl é tipo Titanic: 'Cuidado com os icebergs... ou com a borda.'",
		"💪 *Nadar costas é meu Náufrago: 'Sempre olhando pro horizonte sem fim.'",
		"🏋️‍♂️ *'Minha técnica de borboleta? Mais 'Tubarão' do que 'Grace Kelly'.'",
		"💦 *Nadar peito é como Moana: 'O oceano sempre encontra um jeito de te parar.'",
		"🏊‍♂️ *Treino de fundo é 'Peixe Grande': você se sente uma lenda, mas termina como isca.",
		"🦾 'Sou o rei da piscina!' - até que a criança na raia ao lado me ultrapassa.",
		"💪 *Nado medley é meu Pequena Sereia: 'Cada braçada parece um câmbio entre mundos.'",
		"🏊‍♀️ *Treino de crawl é como Tubarão: 'A água te persegue e você tenta sobreviver.'",
		"💦 *Nadar borboleta é meu 'Avatar 2': imersivo, lindo, mas quase impossível.",
		"🏋️‍♂️ *Treinar em mar aberto é tipo Piratas do Caribe: 'Nunca confie na água.'",
		"💪 *Cada virada é Titanic: 'Você afunda ou vira o rei do mundo (na borda).'",
		"🏊‍♂️ 'Nade mais rápido, eles estão atrás de você!' - Pensamento inspirado em Tubarão.",
		"🦾 *Minha técnica de crawl é tipo Bob Esponja: 'Descoordenada e com bolhas.'",
		"💦 *Depois de nadar 500m, eu me sinto 'Procurando Nemo': totalmente perdido.",
		"🏊‍♀️ 'O que o mar te deu, o treino te tira.' - Filosofia de Moana em dias ruins.",
		"💪 *Cada treino na piscina é como Tubarão: 'Só quero sair vivo no final.'",
		"🏋️‍♂️ *Nadar costas é 'A Forma da Água': estiloso, mas só se você souber o truque.",
		"💦 *Minha resistência na piscina é 'Aquaman': puro marketing, mas sem superpoderes.",
		"🏊‍♂️ 'A piscina tem 25 metros, mas parece o Triângulo das Bermudas.'",
		"🦾 *Nadar é como Moana: 'Você quer atravessar o oceano, mas ele sempre vence.'",
		"💪 *Borboleta é meu Titanic: 'A segunda série sempre afunda.'",
		"🏋️‍♂️ *Cada virada na borda é como Náufrago: 'Você sente que perdeu tudo, menos a touca.'",
		"💦 *Treinar nado costas é meu Pequena Sereia: 'Sempre querendo ar, mas só vejo água.'",
		"🏊‍♀️ 'Eu sou Groot!' - Eu, na borda, tentando explicar o cansaço pro treinador.",
	},
	WeightTraining: {
		"💪 Um supino para todos governar, um agachamento para achá-los, um levantamento terra para a todos trazer e na hipertrofia prendê-los.",
		"🏋️‍♂️ Você não fala sobre o clube da luta, mas todo mundo sabe quando você bate PR no deadlift.",
		"🦵 Expecto Patronum! Porque depois do treino de perna só um feitiço me salva.",
		"🏋️‍♀️ Central Perk? Não, é Central PR: o lugar onde Ross nunca perde as pernas.",
		"💪 *No pain, no gain. Ou, como diria Gandalf: 'You shall not PASS... sem uma boa série de agachamentos!'",
		"🏃‍♂️ Corrida? Isso é muito 'Parkour!', disse Michael Scott enquanto fugia do treino de pernas.",
		"💥 Na academia, eu sou inevitável, igual ao Thanos no leg press.",
		"🦾 Série de bíceps: 'I'll be back.' - Arnold (e você no espelho da academia).",
		"🥵 Treinar com calor é tipo 'Dracarys!' no bíceps. Só falta o dragão do Khaleesi pra me ajudar a respirar.",
		"💪 *Treinamento funcional é tipo Star Wars: 'Que a força esteja com você, mas sem machucar as costas.'",
		"🏋️‍♂️ 'Stairway to Heaven'? Não, é a escada infinita da academia e eu não vejo o céu, só o suor.",
		"🦵 Treino de pernas é como a Caverna do Dragão: você nunca acha a saída.",
		"💥 *Sou tipo o Hulk no supino: quanto mais bravo fico, mais levanto. 'Smash!'",
		"🤔 Os amigos falam 'Pivote! Pivote!' enquanto eu tento erguer o halter mais pesado.",
		"🚴‍♂️ Subir no spinning é o meu 'Winter is coming.' Só que o inverno sou eu sofrendo.",
		"🦾 *Quando alguém rouba meu aparelho: 'Avengers... Assemble! No meu horário!.'",
		"🦵 Depois de um treino de perna, eu me sinto 'Um Jedi caído'. E o sabre? Minha toalha molhada.",
		"🏋️‍♀️ Treinar tríceps é como Friends: 'They don’t know that we know that they know!' Mas eu sei que dói.",
		"🦾 O levantamento terra não é 'Stranger Things', mas me manda direto pro mundo invertido.",
		"🏋️‍♂️ *Treino de bíceps: 'Vou fazer isso o dia todo.' - Capitão América enquanto segura o halter.",
		"💪 'Ninguém faz a menor ideia do peso que eu carrego... porque eu treino sozinho.' - Inspirado em Dark.",
		"🏋️‍♂️ 'Meu precioso! - Disse eu para o halter de 50kg no levantamento terra.' - Gollum mode on.",
		"🦵 *Treino de perna é tipo Matrix: 'There is no spoon, só dor.'",
		"💥 'Com grandes PRs vêm grandes responsabilidades.' - O mantra do Homem-Aranha na academia.",
		"🏃‍♂️ A esteira é como Jurassic Park: quanto mais rápido você corre, mais parece que algo tá te caçando.",
		"🦾 *Quando vejo alguém roubando meu banco no supino: 'Say my name!' - Walter White mode ativado.",
		"🏋️‍♀️ Agachamento é tipo 'O Poderoso Chefão': você sempre paga um preço no final.",
		"🥵 *Treinar no calor é como Mad Max: 'Fury Road' versão academia.",
		"💪 Treino de tríceps: 'I am vengeance.' - Batman, depois de trincar o braço no espelho.",
		"🦵 'Assim é como termina o mundo, não com um estrondo, mas com um treino de perna.' - Poeta e cansado.",
		"🚴‍♂️ 'Me chama de Flash, mas no spinning.' - Disse ninguém, enquanto pedala com 3 watts.",
		"💥 *Deadlift é tipo Transformers: 'Mais do que os olhos conseguem ver' no peso.",
		"🦾 'É perigoso ir sozinho, leve esta toalha!' - Zelda na academia, sempre prevenido.",
		"🏋️‍♂️ *Subir no rack de agachamento é tipo Interstellar: 'O tempo passa diferente lá dentro.'",
		"💪 Treino de costas? Chame de 'Breaking Bad': porque o trapézio não mente.",
		"🦵 *Treino de perna é como Stranger Things: você se sente no 'Upside Down' logo no segundo exercício.",
		"🥵 'Até que os ventos do Sahara soprem... ou que o ventilador da academia funcione.' - Inspirado em Aladdin.",
		"🏋️‍♀️ 'May the PRs be with you.' - Star Wars do agachamento.",
		"🦾 *Depois do cardio, é tipo Gladiador: 'Are you not entertained?' - Eu, suando igual Maximus.",
		"🏃‍♂️ A academia depois das festas é como The Walking Dead: só vejo zumbis no leg day.",
		"💪 'Eu posso fazer isso o dia todo.' - Capitão América, também conhecido como seu personal no supino.",
		"🏋️‍♂️ *Treino de ombro é tipo Titanic: você sente o 'Iceberg!' logo no meio da série.",
		"🦵 *Agachamento é o meu 'Círculo de Fogo': luto contra monstros... meus próprios limites!",
		"💥 'Eu sou o perigo.' - Walter White e eu, quando levanto 200kg no terra.",
		"🏃‍♂️ Corrida na esteira é como 'Gravidade': você se sente flutuando, mas é só o suor.",
		"🦾 'Say hello to my little friend.' - Eu e meu halter de 50kg, direto de Scarface.",
		"🥵 'I'm the king of the world!' - Eu, na última repetição de bíceps. Titanic vibes.",
		"🚴‍♂️ *Subir no spinning é tipo De Volta Para o Futuro: 'Onde estamos indo, não precisamos de descanso.'",
		"💪 'Avada Kedavra!' - O feitiço que uso na dor muscular pós-agachamento.",
		"🦵 *Treino de perna é como o Mundo de Avatar: 'Você descobre músculos que nem sabia que existiam.'",
		"💥 'Eu sou Groot.' - Meu mantra enquanto levanto peso no terra.",
		"🏋️‍♀️ *Treino de tríceps é tipo Game of Thrones: 'A dor está vindo.'",
		"🦾 'O que não me mata, me fortalece.' - Batman, enquanto malha o peitoral.",
		"🥵 *Depois de um treino intenso, é tipo 'Interestelar': um minuto no rack, sete anos no chuveiro.",
		"💪 *Treino de costas é como King Kong: 'Só os fortes sobrevivem.'",
		"🏃‍♂️ *Correr na esteira é como Missão Impossível: 'Não olhe para trás, ou você tropeça.'",
		"🦾 'Você levanta, ou morre tentando.' - Meu lema no supino, inspirado em 50 Cent.",
		"💥 *Treino de ombro é como Star Trek: 'Explorando novos limites.'",
		"🦵 'Hakuna Matata!' - Minha filosofia no leg day... até a terceira série.",
		"🏋️‍♂️ *Cada levantamento terra é tipo Jurassic Park: 'Você escuta ossos estalando no fundo.'",
		"💪 'É tudo sobre poder infinito!' - Thanos e eu no leg press.",
		"🚴‍♂️ *Treino de bike é como Forrest Gump: 'Eu só continuei pedalando.'",
		"🦵 *Agachamento é tipo Doctor Who: 'Sempre regenerando a força.'",
		"💥 *Treinar costas é como Os Incríveis: 'Mais trapézio, menos papo.'",
		"🏃‍♂️ 'Run, Forrest, run!' - Meu mantra no cardio de segunda-feira.",
		"🦾 *Levantar peso é tipo Toy Story: *'Há um halter no meu caminho!'",
		"💪 *Treino de bíceps é como O Grande Lebowski: 'Isso amarra tudo junto.'",
		"🥵 *Treino funcional é tipo Stranger Things: 'Você só quer sair do Upside Down.'",
		"🏋️‍♀️ *Cada série de supino é como Breaking Bad: *'Say my PR!'",
		"💥 *Deadlift é meu Matrix pessoal: 'Eu vejo o código... em cada repetição.'",
	},
	Yoga: {
		"🧘‍♀️ 'Hoje eu escolho acreditar em mim mesmo.' - Mas só depois de conseguir sair dessa pose impossível.",
		"🪷 'Inspire paz, expire gratidão.' - E uma boa dose de dor no alongamento.",
		"🧘‍♂️ 'Seja como a água, flua.' - Pena que eu sou mais como concreto: duro e imóvel.",
		"🕉️ 'A energia que você dá ao universo, você recebe de volta.' - Então, por que só volta câimbra?",
		"🧘 'Hoje eu abraço minha jornada.' - Mesmo que ela seja tropeçar no tapetinho.",
		"🪷 'Você é o mestre do seu destino.' - Exceto quando tenta o cachorro olhando para baixo.",
		"✨ 'Confie no processo.' - Eu confio, mas meu quadril não parece estar nessa vibe.",
		"🌱 'Agradeça ao seu corpo pelo que ele pode fazer.' - Ok, corpo, obrigada por reclamar em todas as poses.",
		"🧘‍♀️ 'Aquiete sua mente e ouça seu corpo.' - Ele está gritando: ‘Sai dessa posição!’",
		"🧘‍♂️ 'Encontre sua paz interior.' - Ela provavelmente está escondida no fundo do meu armário de biscoitos.",
		"✨ 'Seja presente no momento.' - Difícil, quando o momento envolve meu nariz grudado no joelho.",
		"🧘‍♀️ 'Respire fundo e solte o que não te serve.' - A gravidade certamente não está ajudando.",
		"🌿 'Cada dia é um novo começo.' - Exceto para minha flexibilidade, que parou nos anos 90.",
		"🌙 'Você é luz, você é amor.' - Mas hoje eu sou só dor nas costas.",
		"🧘‍♂️ 'Aceite o que é e deixe ir.' - Especialmente a ideia de parecer gracioso fazendo yoga.",
		"🌟 'Seu corpo é um templo.' - Um templo em reforma com andaimes caindo.",
		"🌼 'Onde o foco vai, a energia flui.' - Então, por que minha energia flui direto para a desistência?",
		"🧘 'Abra seu coração.' - E provavelmente uma costela, tentando essa torção.",
		"🪷 'Ame a si mesmo completamente.' - Inclusive as partes que odeiam a posição da árvore.",
		"🌺 'A dor é temporária.' - Mas o trauma de tentar aquela inversão vai durar para sempre.",
		"🧘‍♂️ 'Deixe ir o que não serve mais.' - Incluindo minhas expectativas sobre um alongamento decente.",
		"🌙 'Ouça sua respiração.' - Parece mais um motor engasgando, mas tudo bem.",
		"🌿 'Seja gentil consigo mesmo.' - Especialmente quando cair pela quinta vez.",
		"✨ 'Permita-se simplesmente ser.' - Contorcido e confuso na posição da cobra.",
		"🌟 'Tudo acontece por uma razão.' - Inclusive essa dor que eu não sabia que existia.",
		"🧘‍♂️ 'Mente quieta, coração aberto.' - Mas meu quadril está claramente revoltado.",
		"🌺 'Visualize seu melhor eu.' - Ele provavelmente está sentado no sofá, assistindo TV.",
		"🌼 'A paz começa dentro de você.' - E aparentemente termina assim que tento a pose do guerreiro.",
		"🪷 'Permita-se florescer.' - Mesmo que você pareça mais um cacto tentando yoga.",
		"🧘 'Seja como uma folha ao vento.' - Ou como um tronco quando eu caio.",
		"🌿 'A prática te leva à perfeição.' - Ou pelo menos ao ortopedista.",
		"🕉️ 'Aceite sua jornada única.' - Mesmo que ela envolva tropeçar no tapetinho.",
		"🧘‍♀️ 'Encontre sua força interior.' - Provavelmente escondida sob uma montanha de preguiça.",
		"✨ 'Tudo está conectado.' - Inclusive meu ego e a vergonha de cair na aula.",
		"🌙 'Escolha a calma.' - Difícil, quando o instrutor diz que isso era só o aquecimento.",
		"🌼 'Sinta-se grato pelo agora.' - Mesmo que o ‘agora’ envolva dor na lombar.",
		"🌟 'Você é um ser ilimitado.' - Exceto no alongamento, porque ali sou bem limitado.",
		"🧘‍♂️ 'Cada respiração é um renascimento.' - Pena que renasço cansado em todas.",
		"🌿 'O universo está em você.' - Certamente não na parte que entende essa pose invertida.",
		"🌺 'Celebre suas pequenas vitórias.' - Como sobreviver à aula sem ficar preso na pose do pombo.",
	},
	Walk: {
		"🚶‍♂️ *Minha caminhada é tipo Pica-Pau: 'Sorrindo por fora, mas pronto pra correr se o problema aparecer.'",
		"🏃‍♂️ *Caminhar é meu Tom e Jerry: 'Cada passo parece uma fuga de algo invisível.'",
		"🚶‍♀️ 'Caminho tanto que me sinto no Reino dos Cogumelos: só falta o Mario pra me salvar.'",
		"💪 *Caminhar no calor é meu Rick and Morty: 'Sempre em outra dimensão, longe do ar-condicionado.'",
		"🏞️ 'Pegue sua espada, Mestre dos Magos está perto!' - Pensamento recorrente nas subidas.",
		"🚶‍♂️ *Minha caminhada é Bob Esponja: 'Eu tentando parecer animado enquanto tudo que quero é parar.'",
		"💥 *Caminhar ao ar livre é tipo Meninas Super Poderosas: 'Lutando contra o cansaço, o crime... e o sol.'",
		"🚶‍♀️ *Cada passo na subida é Caverna do Dragão: 'O portal pra casa nunca aparece.'",
		"💪 'Vamos caminhar!' - Disse eu, acreditando ser uma Espiã Demais. Spoiler: não sou.",
		"🏃‍♂️ *Caminhar é meu Rick and Morty: 'Sempre acho que cheguei, mas a caminhada continua.'",
		"🚶‍♂️ 'Hora de aventura!' - Até perceber que a subida é longa e o fôlego é curto.",
		"💪 *Caminhar é tipo Tom e Jerry: 'Eu sou o Tom e a ladeira é o Jerry... sempre fugindo de mim.'",
		"🚶‍♀️ 'Preparem-se, amigos!' - Eu no início da caminhada, mas sem o Pikachu pra carregar meu peso.",
		"💥 *Caminhar no frio é tipo Scooby-Doo: 'Sempre correndo de algo imaginário.'",
		"🏃‍♂️ *Caminhar na areia é meu Bob Esponja: 'A cada passo, me sinto mais como o Patrick.'",
		"🚶‍♂️ 'Você não passa!' - Gandalf, ou a subida que enfrento toda semana.",
		"💪 *Cada ladeira é tipo Pica-Pau: 'Ela sobe, eu paro. Ela desiste? Nunca.'",
		"🚶‍♀️ *Minha caminhada é como Dragon Ball Z: 'Parece que nunca chega ao final.'",
		"💥 *Caminhar em círculos é meu Caverna do Dragão: 'Você nunca encontra a saída.'",
		"🚶‍♂️ 'Eu sou invencível!' - Até o primeiro morro acabar com meu ânimo.",
		"💪 *Minha caminhada é Scooby-Doo: *'Você resolve um mistério a cada passo: ‘Onde foi parar minha energia?’",
		"🚶‍♀️ *Caminhar ouvindo música é tipo Rick and Morty: 'Uma nova realidade com cada música que toca.'",
		"🏃‍♂️ 'Preparem-se para a próxima caminhada!' - Eu, tentando ser James do Team Rocket na subida.",
		"🚶‍♂️ *Caminhar é meu Hora de Aventura: 'Lutando contra o sono, as subidas e o cansaço.'",
		"💪 *Cada passo é tipo Tom e Jerry: 'Você está sempre tentando pegar algo, mas nunca alcança.'",
		"🚶‍♀️ *Caminhar na chuva é meu Bob Esponja: 'Só falta a música triste pra completar.'",
		"💥 *Cada subida é Pica-Pau: 'Uma risada debochada na minha cara enquanto eu sofro.'",
		"🏃‍♂️ *Caminhar rápido é como Dragon Ball Z: 'Tudo parece em câmera lenta enquanto o suor aumenta.'",
		"🚶‍♂️ 'Eu tenho o poder!' - Disse ninguém ao enfrentar o morro mais íngreme.",
		"💪 *Caminhar no parque é Scooby-Doo: 'Você vê sombra, mas jura que é um monstro.'",
		"🚶‍♀️ 'Essa ladeira é como o Mestre dos Magos: aparece do nada e te faz sofrer.'",
		"💥 *Caminhar com amigos é tipo Rick and Morty: 'Cada conversa é uma viagem interdimensional.'",
		"🚶‍♂️ 'Caminhar é um grande mistério.' - Scooby-Doo enquanto investiga meus passos lentos.",
		"💪 *Cada subida é Tom e Jerry: 'Você tenta vencer, mas só toma rasteira.'",
		"🚶‍♀️ *Caminhar é como Pokémon: 'Você só avança se tiver uma poção no bolso (garrafa d’água).'",
		"💥 *Caminhar com mochila é meu Bob Esponja: 'Carregando tudo, menos força.'",
		"🏃‍♂️ *Cada volta no parque é Rick and Morty: 'Parece infinito, mas é só o cansaço te enganando.'",
		"🚶‍♂️ *Caminhar ao sol é como Meninas Super Poderosas: 'Sofrendo, mas sempre estilosas.'",
		"💪 *Caminhada é tipo Dragon Ball Z: 'Só gritando você acredita que vai chegar.'",
		"🚶‍♀️ 'Continue andando!' - O Mestre dos Magos enquanto ignora meu pedido de descanso.",
		"🚶‍♂️ Caminhando na velocidade do Internet Explorer",
		"🚶‍♀️ Andando mais que Pokémon sem Pokébola",
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
