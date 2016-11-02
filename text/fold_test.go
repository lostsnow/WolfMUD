// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package text

import (
	"bytes"
	"fmt"
	"testing"
)

var testCases = []struct {
	data  string
	width int
	want  string
}{
	// Basic string folding on different widths
	{"WolfMUD - World Of Living Fantasy", 35, "WolfMUD - World Of Living Fantasy"},
	{"WolfMUD - World Of Living Fantasy", 33, "WolfMUD - World Of Living Fantasy"},
	{"WolfMUD - World Of Living Fantasy", 30, "WolfMUD - World Of Living\r\nFantasy"},
	{"WolfMUD - World Of Living Fantasy", 20, "WolfMUD - World Of\r\nLiving Fantasy"},
	{"WolfMUD - World Of Living Fantasy", 15, "WolfMUD - World\r\nOf Living\r\nFantasy"},
	{"WolfMUD - World Of Living Fantasy", 10, "WolfMUD -\r\nWorld Of\r\nLiving\r\nFantasy"},
	{"WolfMUD - World Of Living Fantasy", 9, "WolfMUD -\r\nWorld Of\r\nLiving\r\nFantasy"},
	{"WolfMUD - World Of Living Fantasy", 8, "WolfMUD\r\n- World\r\nOf\r\nLiving\r\nFantasy"},
	{"WolfMUD - World Of Living Fantasy", 7, "WolfMUD\r\n- World\r\nOf\r\nLiving\r\nFantasy"},
	{"WolfMUD - World Of Living Fantasy", 6, "WolfMUD\r\n-\r\nWorld\r\nOf\r\nLiving\r\nFantasy"},
	{"WolfMUD - World Of Living Fantasy", 5, "WolfMUD\r\n-\r\nWorld\r\nOf\r\nLiving\r\nFantasy"},
	{"WolfMUD - World Of Living Fantasy", 4, "WolfMUD\r\n-\r\nWorld\r\nOf\r\nLiving\r\nFantasy"},
	{"WolfMUD - World Of Living Fantasy", 3, "WolfMUD\r\n-\r\nWorld\r\nOf\r\nLiving\r\nFantasy"},
	{"WolfMUD - World Of Living Fantasy", 2, "WolfMUD\r\n-\r\nWorld\r\nOf\r\nLiving\r\nFantasy"},
	{"WolfMUD - World Of Living Fantasy", 1, "WolfMUD\r\n-\r\nWorld\r\nOf\r\nLiving\r\nFantasy"},
	{"WolfMUD - World Of Living Fantasy", 0, "WolfMUD - World Of Living Fantasy"},
	{"WolfMUD - World Of Living Fantasy", -1, "WolfMUD - World Of Living Fantasy"},
	{"WolfMUD - World Of Living Fantasy", -10, "WolfMUD - World Of Living Fantasy"},
	{"WolfMUD - World Of Living Fantasy", -20, "WolfMUD - World Of Living Fantasy"},
	{"WolfMUD - World Of Living Fantasy", -30, "WolfMUD - World Of Living Fantasy"},

	// Originally caused a panic before we stopped wrapping on widths less than 1
	{"WolfMUD - World Of Living Fantasy", -40, "WolfMUD - World Of Living Fantasy"},

	// Leading and/or trailing whitespace
	{"   WolfMUD", 13, "   WolfMUD"},
	{"   WolfMUD", 10, "   WolfMUD"},
	{"   WolfMUD", 7, "   WolfMUD"},
	{"   WolfMUD", 5, "   WolfMUD"},
	{"WolfMUD   ", 13, "WolfMUD"},
	{"WolfMUD   ", 10, "WolfMUD"},
	{"WolfMUD   ", 7, "WolfMUD"},
	{"WolfMUD   ", 5, "WolfMUD"},
	{"   WolfMUD   ", 13, "   WolfMUD"},
	{"   WolfMUD   ", 10, "   WolfMUD"},
	{"   WolfMUD   ", 7, "   WolfMUD"},
	{"   WolfMUD   ", 5, "   WolfMUD"},

	// Multi-line input cases
	{"\nWolfMUD", 10, "\r\nWolfMUD"},
	{"WolfMUD\n", 10, "WolfMUD\r\n"},
	{"\nWolfMUD\n", 10, "\r\nWolfMUD\r\n"},
	{"WolfMUD\nWolfMUD", 10, "WolfMUD\r\nWolfMUD"},
	{"WolfMUD\n\nWolfMUD", 10, "WolfMUD\r\n\r\nWolfMUD"},
	{"WolfMUD\n\n\nWolfMUD", 10, "WolfMUD\r\n\r\n\r\nWolfMUD"},
	{"\n\nWolfMUD\nWolfMUD", 10, "\r\n\r\nWolfMUD\r\nWolfMUD"},
	{"WolfMUD\nWolfMUD\n\n", 10, "WolfMUD\r\nWolfMUD\r\n\r\n"},

	{"\nWolfMUD", 8, "\r\nWolfMUD"},
	{"WolfMUD\n", 8, "WolfMUD\r\n"},
	{"\nWolfMUD\n", 8, "\r\nWolfMUD\r\n"},
	{"WolfMUD\nWolfMUD", 8, "WolfMUD\r\nWolfMUD"},
	{"WolfMUD\n\nWolfMUD", 8, "WolfMUD\r\n\r\nWolfMUD"},
	{"WolfMUD\n\n\nWolfMUD", 8, "WolfMUD\r\n\r\n\r\nWolfMUD"},
	{"\n\nWolfMUD\nWolfMUD", 8, "\r\n\r\nWolfMUD\r\nWolfMUD"},
	{"WolfMUD\nWolfMUD\n\n", 8, "WolfMUD\r\nWolfMUD\r\n\r\n"},

	{"\nWolfMUD", 7, "\r\nWolfMUD"},
	{"WolfMUD\n", 7, "WolfMUD\r\n"},
	{"\nWolfMUD\n", 7, "\r\nWolfMUD\r\n"},
	{"WolfMUD\nWolfMUD", 7, "WolfMUD\r\nWolfMUD"},
	{"WolfMUD\n\nWolfMUD", 7, "WolfMUD\r\n\r\nWolfMUD"},
	{"WolfMUD\n\n\nWolfMUD", 7, "WolfMUD\r\n\r\n\r\nWolfMUD"},
	{"\n\nWolfMUD\nWolfMUD", 7, "\r\n\r\nWolfMUD\r\nWolfMUD"},
	{"WolfMUD\nWolfMUD\n\n", 7, "WolfMUD\r\nWolfMUD\r\n\r\n"},

	{"\nWolfMUD", 6, "\r\nWolfMUD"},
	{"WolfMUD\n", 6, "WolfMUD\r\n"},
	{"\nWolfMUD\n", 6, "\r\nWolfMUD\r\n"},
	{"WolfMUD\nWolfMUD", 6, "WolfMUD\r\nWolfMUD"},
	{"WolfMUD\n\nWolfMUD", 6, "WolfMUD\r\n\r\nWolfMUD"},
	{"WolfMUD\n\n\nWolfMUD", 6, "WolfMUD\r\n\r\n\r\nWolfMUD"},
	{"\n\nWolfMUD\nWolfMUD", 6, "\r\n\r\nWolfMUD\r\nWolfMUD"},
	{"WolfMUD\nWolfMUD\n\n", 6, "WolfMUD\r\nWolfMUD\r\n\r\n"},

	// Weird corner cases
	{"", 0, ""},
	{"", 1, ""},
	{"", 2, ""},
	{"\n", 0, "\r\n"},
	{"\n", 1, "\r\n"},
	{"\n", 2, "\r\n"},

	// UTF-8 2 bytes 0xc2 0xa3, Unicode U+00A3, £, POUND SIGN
	{"Unicode \u00A3", 9, "Unicode \u00A3"},

	// UTF-8 3 bytes 0xe2 0x88 0x91, Unicode U+2211, ∑, N-ARY SUMMATION
	{"Unicode \u2211", 9, "Unicode \u2211"},

	// UTF-8 4 bytes 0xf0 0x9f 0x9e 0x8e, Unicode U+1F78E, 🞎, LIGHT WHITE SQUARE
	{"Unicode \U0001f78e", 9, "Unicode \U0001f78e"},

	// Combining characters: LATIN SMALL LETTER A + COMBINING GRAVE ACCENT
	{"Unicode \u0061\u0300", 9, "Unicode \u0061\u0300"},

	// Control sequences should be zero width
	{"\033[31mWolfMUD\033[39m", 7, "\033[31mWolfMUD\033[39m"},
	{"\n\033[31mWolfMUD\033[39m\n", 7, "\r\n\033[31mWolfMUD\033[39m\r\n"},
	{"\033[31mWolfMUD WolfMUD\033[39m", 20, "\033[31mWolfMUD WolfMUD\033[39m"},
	{"\n\033[31mWolfMUD WolfMUD\033[39m\n", 20, "\r\n\033[31mWolfMUD WolfMUD\033[39m\r\n"},
	{"\n\033[31mWolfMUD WolfMUD\033[39m\n", 20, "\r\n\033[31mWolfMUD WolfMUD\033[39m\r\n"},
	{"\033[31mWolfMUD \033[32mWolfMUD\033[39m", 7, "\033[31mWolfMUD\r\n\033[32mWolfMUD\033[39m"},
	{"\033[31mWolfMUD\033[32mWolfMUD\033[39m", 7, "\033[31mWolfMUD\033[32mWolfMUD\033[39m"},
	{"\033[31mWolfMUD\n\033[32mWolfMUD\033[39m", 7, "\033[31mWolfMUD\r\n\033[32mWolfMUD\033[39m"},
	{"\033[31mWolfMUD\n \033[32mWolfMUD\033[39m", 7, "\033[31mWolfMUD\r\n \033[32mWolfMUD\033[39m"},
	{"\033[31mWolfMUD\n\033[32m WolfMUD\033[39m", 7, "\033[31mWolfMUD\r\n\033[32m WolfMUD\033[39m"},
	{"\033[31mWolfMUD\n  \033[32mWolfMUD\033[39m", 7, "\033[31mWolfMUD\r\n  \033[32mWolfMUD\033[39m"},
	{"\033[31mWolfMUD\n\033[32m  WolfMUD\033[39m", 7, "\033[31mWolfMUD\r\n\033[32m  WolfMUD\033[39m"},
}

func TestFold(t *testing.T) {
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Fold width %d", tc.width), func(t *testing.T) {
			have := Fold([]byte(tc.data), tc.width)
			if !bytes.Equal(have, []byte(tc.want)) {
				t.Errorf("\nhave %+q\nwant %+q", have, tc.want)
			}
		})
	}
}

var lipsumASCII = `Etiam sed arcu at tellus volutpat porta. Aenean gravida hendrerit
gravida. Cras vehicula vestibulum lacus sit amet convallis. Integer rutrum
auctor nibh, vitae volutpat sem hendrerit vestibulum. Quisque volutpat
sollicitudin tortor vel posuere. Duis lacus metus, mattis et cursus eu, euismod
ullamcorper magna. Aenean erat nisl, tempus at lobortis sit amet, sollicitudin
id augue. Sed eget mauris ante. Vestibulum a mattis tortor. Duis sit amet eros
in enim aliquam consectetur.

Curabitur egestas nunc vitae tortor consequat dignissim. Ut ac dui felis. Nulla
fringilla, lectus id ultrices sagittis, tellus massa facilisis libero, at porta
mauris tortor vel turpis. Donec tempus, metus in molestie fermentum, massa nibh
malesuada nulla, eu dignissim risus arcu quis lacus. Nunc massa ipsum, blandit
non viverra a, scelerisque vel eros. Duis congue malesuada massa ut
sollicitudin. Curabitur lacinia aliquam elementum. Nam elementum scelerisque
laoreet. Pellentesque et lectus velit. Fusce convallis purus non eros imperdiet
id ullamcorper erat varius. Duis iaculis dui a metus mattis vitae ullamcorper
erat semper. Cras orci mauris, sagittis non bibendum id, auctor quis risus.
Suspendisse risus felis, egestas non pulvinar eget, pulvinar sit amet dui.
Fusce vitae arcu dui. Integer nec enim id nisi sollicitudin fermentum. Sed
risus lorem, egestas eu dictum id, adipiscing vel lectus.

In quis imperdiet tortor. Aenean consectetur interdum diam ut rhoncus. Maecenas
posuere, nisi id luctus euismod, odio odio blandit erat, eget aliquam velit
massa non sem. Curabitur vestibulum dignissim purus, non viverra lorem laoreet
varius. Ut hendrerit augue eu leo vehicula vel porta libero facilisis.
Phasellus volutpat tortor in nisi dictum vel dictum massa interdum. Ut mauris
dui, fringilla quis tincidunt vitae, faucibus pretium mi. Nunc lobortis
interdum imperdiet. Donec id rutrum diam. Vestibulum vel augue mauris, et
venenatis enim. Suspendisse consequat erat ut nisl volutpat ultricies nec vitae
enim. Quisque sodales quam ut quam pharetra egestas. Pellentesque sed odio sem.

Mauris hendrerit, purus et dapibus tristique, metus dolor semper diam, ac
congue tortor tortor vel erat. Vestibulum dictum arcu at ligula molestie
tincidunt. Donec ac dui vitae ante sodales imperdiet. Nulla facilisi. Donec
adipiscing pulvinar nulla. Curabitur et elit erat, sit amet vehicula mi. Morbi
varius vulputate condimentum. Curabitur vestibulum lacus in tellus varius
blandit. Integer fringilla faucibus porta. Duis imperdiet libero a tortor
cursus vel auctor nulla commodo. Mauris commodo, lorem ac condimentum interdum,
quam urna tempor felis, in ornare diam erat id felis.

Integer pellentesque ultrices facilisis. Sed eget odio non sem feugiat rutrum.
Etiam pharetra imperdiet tristique. Donec posuere arcu quis justo molestie
posuere id eu tellus. Integer posuere dignissim justo et mollis. Nullam
ultricies sem sit amet ipsum facilisis volutpat. Ut viverra nulla a tellus
adipiscing congue. In hac habitasse platea dictumst. Sed blandit egestas est,
ac porttitor nisl feugiat eu. Sed gravida lacus sit amet turpis fringilla ac
malesuada magna porttitor. Praesent vitae neque eget orci accumsan lobortis.
Morbi dignissim tincidunt nisl, at convallis purus suscipit at. Nunc sed risus
scelerisque augue posuere gravida. Cras ac massa felis.

Phasellus cursus, arcu ut lacinia pulvinar, ipsum erat scelerisque diam, nec
condimentum mi eros in risus. Nunc hendrerit accumsan sapien porttitor
fringilla. Nunc id quam velit, ut commodo nisl. Integer eget lobortis neque.
Cras vel gravida nunc. Aliquam cursus sodales lectus ut faucibus. Nam nec
semper quam. Nullam a nibh et orci placerat gravida ut id leo. Maecenas lacus
sem, placerat vel blandit id, tincidunt vel odio. Maecenas ac tortor tortor,
vitae eleifend augue. Proin adipiscing tortor non lectus auctor vitae
adipiscing nulla sagittis. Vestibulum pharetra quam vehicula ante vestibulum
sodales. Mauris sit amet mauris enim.

Nullam a turpis risus, non fermentum tortor. Maecenas gravida sagittis
fermentum. Quisque id nisl sem, ut euismod eros. Duis purus purus, pharetra sed
varius vel, bibendum ac sapien. Lorem ipsum dolor sit amet, consectetur
adipiscing elit. Proin volutpat tellus ut eros euismod congue. Cras est mauris,
facilisis a volutpat id, hendrerit ac risus. Aliquam vel turpis tortor, a
pellentesque massa. In commodo semper mattis. Vestibulum dapibus, magna non
volutpat gravida, velit lorem sodales dui, nec rutrum est libero non felis.
Proin leo dui, sodales vel semper a, rhoncus quis sapien. Pellentesque
condimentum feugiat sem ac posuere. Morbi semper nibh in massa placerat eu
ullamcorper est sollicitudin. Cras pulvinar magna egestas felis faucibus et
tristique magna condimentum. Vestibulum eget lobortis nisi. Sed id tincidunt
urna. Nam mi elit, posuere eget scelerisque ac, lobortis non nisl.`

func BenchmarkFoldLipsumASCII(b *testing.B) {
	text := []byte(lipsumASCII)
	for _, width := range []int{20, 40, 80, 100, 120, 140, 160} {
		b.Run(fmt.Sprintf("Width %d", width), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = Fold(text, 80)
			}
		})
	}
}

var lipsumUTF8 = `Sia mo plej traigi emfazado, neado patro minusklo ant ik. Os oho
fojo cento pleja. In frazo antaŭmetado ind. Eko aliio depost ut, kiomas
konsonanto se ree. Am ses jokto nigra nomial, persa poŭpo tempolongo sub os, li
vic itismo kapabl. Om tre vela deci. Se nano nekutima prepozitivo ind. Em fari
tial definitive iel, mil plus perlabori uk. Tie resti semajntago sensubjekta
li, ajna festonomo eko em. Hekto pleje mezurunuo hot um, nenia mekao unu ig,
faka kristnasko gv nul.

Io kuo tiom kontraŭa, des be dura tiaĵo nominativa, no eca vela kompleksa
malantaŭe. Dio is aligi nederlando, tra termo afrikato duonvokalo er, ki kapabl
primitiva jen. Ism retro ekskluzive mallongigoj ok, iufoje ricevanto
frazetvorto sor as. Alie duondifina sin em, ko ofon estro kia. Dua nj vira
esperantigita. Go nulo lasi frazospeco hav, dum fari dank' turpa id, gh alta
predikato reprezenti jam. Aga sube senforte li, eg tek dume latino malantaŭ.
Nedifinita alimaniere ioj ju, ajn mf ekoo kial, san poste responde ec. Am imaga
kompleksa kun. Ni pli verbo ologi estiel, aĥ turpa sekve kilogramo ot, veo
supersigno ligfinaĵo iv. Kaŭze laringalo aga la, hot verba komparado bv, bio
fare parentezo elnombrado e. At ien trudi nette ekesti, ali ar vato decimaloj.
Pli jeso ilion priskribo os. Getto punktokomo eksteraĵo ba sat, jugoslavo
anstataŭe ba tro. For onjo nome ts, do suba viro grupo.

Eŭro esperanteca ehe po, trafe glota fiksi gv duo. Tempa hebrea duondifina vir
at, ial mi land fiksi mallongigita. Lasi kovri emfazo in iom. Ar nia dume neni
ologi. Ne kaj jena kiel antaŭtagmezo. Olog eĥo kompreneble os vir. Ol disde
adjektiva per. Ho dikfingro neŭtrala reprezenti ist. Ekkria substantiva fin io,
el nei triliono prapostmorgaŭ. Per brosi kontraŭi ok. Primitiva samideano
geinstruisto sia la, sub deka kunskribado ro, tuja dupunkto tia ad. Giga
pantalono vortfarado nia if. Video elrigardi e tiu. Far it estr futuro
internacia. Super transigi nekutima so sed, rilate deziri predikativo tc esk.
Plie jesi perlabori dis ne. Ek tria olog eligi ioj. Mano alta io pli, ojd dume.

Senforte sensubjekta ies el, malantaŭ kondicionalo mal bo. Ar jesa alikaŭze
substantiva log, sep ts elrigardi kromakcento. Ie meta respondeci haltostreko
ien, op spite numeralo interalie ojd. Unt kelk fine faras uj, as ree nula
substantiva. Lo depost landonomo ato, om ing reala hosana. Us mili iliard
senobjekta ena. Inter intere tabelvorto ve nur, ekde stop naŭ lo sed. Lo horo
ekster iam, ree id ador mega duuma. In ali tera tele centi. Halo bedaŭrinde ik
pli. Pleja volitivo malantaŭa vi veo, ing suba vasta nekutima po. Dura grupo
inkluzive nei ni, kiom laŭlonge miriametro fo ian.

Ge longa frazospeco indikativo tri, ligvokalo alternativo be kio. Nv peti kurta
futuro dek, nome nenii ek dis. Cii ec konateco tiudirekten, kia po duon malcit
prapostmorgaŭ. Unuo nenia la plue, timi egalo poŭpo dis do. Ari mi foren multa,
duonvokalo interjekcio afganistano on tri. Nei iu triangulo alternativa
dividostreko. Nk ind kial fiksa kiomas. Vira zorgi reciprokeco ido er, fri in
kunmetita tripunkto kompleksa. Iufoje nuancado iz pre, ato tiea interalie
elnombrado mo, ik anc ekoo mono popolnomo. Jam id lanta tempodaŭro, definitive
antaŭtagmezo ili bo. Aliio nomial ts sub, subigi malprofitanto pseŭdoafikso ed
sis. Meze ekkrio antaŭparto ko vir. Un' nepo komplika asterisko.

Eko fo futuro alternativa frazenkondukilo. Ism ba filo giga neŭtrala. Enz tuje
alii ekkria fo, oj aga vice matematika. Ioj olog duobla si. Et kaj volu
rilativa. Ro eca sori duto. Ehe subfrazo duonvokalo ed. Dume unua komo tet ha,
ut gibi posta tuj. Tia tuja dekono ve, kunmetita finnlando tro fo. Ore ve tiom
finno frakcistreko. Vic en hura poezio. Ec pere kibi helpa pli, pebi cento
leteri be ebl. Afro dikfingro alparolato aŭ ot, kio os kvar voli. Apostrofo
rolmontrilo ke anc, sep dato aliam fundamenta du. Mil fi suba grupo neŭtrala,
pri as meta fini kontraŭa. Nei aliu pleje ie, jo kvanta frazelemento ali, int
peto malsuprenstreko ul.

Nk traigi sezononomo por. Ehe reen dividostreko pseŭdoafikso tc, alia alial
morgaŭo ge ene. Eks re ologi triliono malantaŭ. Bv tempolongo kondicionalo nek.
Sor sh intere negativaj, ok kasedo prirespondi kondicionalo ian. Ok gibi ilion
subjekta cit, vol em verbo semajntago. Plej solstariva vortfarado tc poa, pro
ng ruli helpa tempodaŭro. Aga gibi futuro eksterna o, fo tial tagnokto
okulvitroj nei. Ej neniu postparto festonomo aĥ, tri ioma eŭro festo e. Tri
oble jota oz, aŭ nevo kromakcento ad. Ioj eĥo pebi ek, zo tio jesi certa. Tipo
maldekstre ut kaj, nek ot unua matematiko. Vatto intera el vol. Tiam okej'
franjo ro kiu. Ator postpostmorgaŭ zo pri, onklo parentezo kv end, ebleco
iufoje postmorgaŭ sat fo. Plue nv deksesuma asterisko. Ke tohuo koruso subtraho
hav, oid mi tuja hebrea. Tipo decimala oktiliono li obl, mf fri ikso
konjunkcio. Sia certa multekosta refleksiva si. Tempa afganistano do.`

func BenchmarkFoldLipsumUTF8(b *testing.B) {
	text := []byte(lipsumUTF8)
	for _, width := range []int{20, 40, 80, 100, 120, 140, 160} {
		b.Run(fmt.Sprintf("Width %d", width), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = Fold(text, 80)
			}
		})
	}
}
