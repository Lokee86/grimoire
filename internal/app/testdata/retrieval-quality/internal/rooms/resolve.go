package rooms

// ResolveRoom assigns a player to a room. It is unrelated to combat damage.
func ResolveRoom(playerID string) string {
	return "room:" + playerID
}
