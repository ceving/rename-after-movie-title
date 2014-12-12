package main

type LangId int
type TextId int

const (
	rename_after_movie_title TextId = iota
	rename_the_following_directories
	can_not_find_any_movies
	// Aliases
	application_title = rename_after_movie_title
)

const (
	en LangId = iota
	de
)

var lang LangId = en

var l15n = map[LangId]map[TextId]string {
	en: {
		rename_after_movie_title: "Rename after movie title",
		rename_the_following_directories: "Rename the following directories?",
		can_not_find_any_movies: "Can not find any movies.",
	},
	de: {
		rename_after_movie_title: "Umbenennung gemäß des Film-Titels",
		rename_the_following_directories: "Die folgenden Verzeichnisse umbenennen?",
		can_not_find_any_movies: "Kann keine Filme finden.",
	},
}
