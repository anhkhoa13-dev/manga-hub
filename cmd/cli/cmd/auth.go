package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

var username string
var password string

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to your MangaHub account",
	Run: func(cmd *cobra.Command, args []string) {
		// Gọi HTTP API Login
		reqBody, _ := json.Marshal(map[string]string{
			"username": username,
			"password": password,
		})

		resp, err := http.Post("http://localhost:8080/auth/login", "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			fmt.Println("❌ Lỗi kết nối đến server:", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Println("❌ Đăng nhập thất bại. Vui lòng kiểm tra lại tài khoản/mật khẩu.")
			return
		}

		var result map[string]string
		json.NewDecoder(resp.Body).Decode(&result)

		// Lưu token
		if token, ok := result["token"]; ok {
			saveToken(token)
			fmt.Printf("✓ Login successful!\nWelcome back, %s!\nReady to use MangaHub!\n", username)
		}
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(loginCmd)

	loginCmd.Flags().StringVarP(&username, "username", "u", "", "Your username")
	loginCmd.Flags().StringVarP(&password, "password", "p", "", "Your password")
	loginCmd.MarkFlagRequired("username")
	loginCmd.MarkFlagRequired("password")
}