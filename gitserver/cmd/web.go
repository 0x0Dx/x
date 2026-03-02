package cmd

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/spf13/cobra"

	"github.com/0x0Dx/x/gitserver/routers"
	"github.com/0x0Dx/x/gitserver/utils"
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Run web server",
	Args:  cobra.ArbitraryArgs,
	Run: func(cc *cobra.Command, args []string) {
		fmt.Printf("%s\n", utils.Cfg.App.Name)

		r := chi.NewRouter()
		r.Use(middleware.Logger)
		r.Use(middleware.Recoverer)

		r.Get("/", routers.Dashboard)
		r.Get("/user/signin", routers.SignIn)
		r.Method(http.MethodGet, "/user/signup", http.HandlerFunc(routers.SignUp))
		r.Method(http.MethodPost, "/user/signup", http.HandlerFunc(routers.SignUp))

		listenAddr := fmt.Sprintf("%s:%s",
			utils.Cfg.Server.HTTPAddr,
			utils.Cfg.Server.HTTPPort)
		fmt.Printf("Listen: %s\n", listenAddr)
		http.ListenAndServe(listenAddr, r)
	},
}

func init() {
	RootCmd.AddCommand(webCmd)
}
