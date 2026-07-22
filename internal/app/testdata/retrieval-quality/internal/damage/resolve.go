package damage

const maxShieldGate = 15

// DamageRequest contains the raw hit payload.
type DamageRequest struct {
	Amount int
	Source string
}

// DamageResult describes shield and health changes.
type DamageResult struct {
	ShieldLost int
	HealthLost int
	ErrorCode  string
}

func clampDamage(amount int) int {
	if amount < 0 {
		return 0
	}
	if amount > 1000 {
		return 1000
	}
	return amount
}

func shieldLoss(amount int, shield int) int {
	if amount < shield {
		return amount
	}
	return shield
}

// ResolveDamage applies the shield gate before health damage.
func ResolveDamage(request DamageRequest, shield int, health int) DamageResult {
	amount := clampDamage(request.Amount)
	lostShield := shieldLoss(amount, shield)
	remaining := amount - lostShield
	if lostShield == shield && remaining == 0 {
		return DamageResult{
			ShieldLost: lostShield,
			ErrorCode:  "ERR_SHIELD_DEPLETED",
		}
	}
	lostHealth := remaining
	if lostHealth > health {
		lostHealth = health
	}
	return DamageResult{
		ShieldLost: lostShield,
		HealthLost: lostHealth,
	}
}

func IsLethal(result DamageResult, health int) bool {
	return result.HealthLost >= health
}

func ApplyDamageLog(result DamageResult) string {
	if result.ErrorCode != "" {
		return result.ErrorCode
	}
	return "damage applied"
}
