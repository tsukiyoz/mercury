/**
 * @author tsukiyo
 * @date 2023-08-11 1:15
 */

package domain

type User struct {
	Id        int64
	Email     string
	Password  string
	NickName  string
	Biography string
	Birthday  int64
}
