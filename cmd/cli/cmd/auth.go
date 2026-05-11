package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

var (
	username string
	password string
)

var authCmd = &cobra.Command{Use: "auth", Short: "Authentication commands"}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to MangaHub",
	Run: func(cmd *cobra.Command, args []string) {
		reqBody, _ := json.Marshal(map[string]string{"username": username, "password": password})
		resp, err := http.Post("http://localhost:8080/auth/login", "application/json", bytes.NewBuffer(reqBody))
		if err != nil || resp.StatusCode != http.StatusOK {
			fmt.Println("❌ Đăng nhập thất bại.")
			return
		}
		defer resp.Body.Close()
		var result map[string]string
		json.NewDecoder(resp.Body).Decode(&result)
		if token, ok := result["token"]; ok {
			saveToken(token)
			fmt.Printf("✓ Đăng nhập thành công! Chào mừng %s.\n", username)
		}
	},
}

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new account",
	Run: func(cmd *cobra.Command, args []string) {
		reqBody, _ := json.Marshal(map[string]string{"username": username, "password": password})
		resp, _ := http.Post("http://localhost:8080/auth/register", "application/json", bytes.NewBuffer(reqBody))
		if resp.StatusCode == http.StatusCreated {
			fmt.Println("✓ Đăng ký thành công! Hãy dùng 'mangahub auth login' để tiếp tục.")
		} else {
			fmt.Println("❌ Đăng ký thất bại. Username có thể đã tồn tại.")
		}
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout and clear session",
	Run: func(cmd *cobra.Command, args []string) {
		os.Remove(getTokenPath())
		fmt.Println("✓ Đã đăng xuất thành công.")
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check current login status",
	Run: func(cmd *cobra.Command, args []string) {
		if loadToken() != "" {
			fmt.Println("🟢 Bạn đang đăng nhập.")
		} else {
			fmt.Println("🔴 Bạn chưa đăng nhập.")
		}
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(loginCmd, registerCmd, logoutCmd, statusCmd)

	loginCmd.Flags().StringVarP(&username, "username", "u", "", "Your username")
	loginCmd.Flags().StringVarP(&password, "password", "p", "", "Your password")
	registerCmd.Flags().StringVarP(&username, "username", "u", "", "Your username")
	registerCmd.Flags().StringVarP(&password, "password", "p", "", "Your password")
}