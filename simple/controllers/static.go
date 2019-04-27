package controllers

import "webapp1/simple/views"

func NewStatic() *Static {
	return &Static{
		HomeView:    views.NewView("bootstrap", "static/home"),
		ContactView: views.NewView("bootstrap", "static/contact"),
	}
}

type Static struct {
	HomeView    *views.View
	ContactView *views.View
}
