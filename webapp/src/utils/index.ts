const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const usernameRegex = /^[a-zA-Z0-9._-]{3,32}$/;

export function isValidEmail(email: string) {
  return emailRegex.test(email);
}

export function validateUsernameWithError(username: string): {
  isValid: boolean;
  error?: string;
} {
  if (!username) {
    return { isValid: false, error: "Username is required" };
  }

  if (username.length < 3) {
    return { isValid: false, error: "Username must be at least 3 characters" };
  }

  if (username.length > 32) {
    return { isValid: false, error: "Username must not exceed 32 characters" };
  }

  if (!usernameRegex.test(username)) {
    return {
      isValid: false,
      error:
        "Username can only contain letters, numbers, dots, underscores, and hyphens",
    };
  }

  return { isValid: true };
}


export function validatePassword(pw: string): { isValid: boolean; msg: string } {
  // Length check
  if (pw.length < 12) {
    return { isValid: false, msg: "Password must be at least 12 characters long." };
  }

  if (pw.length > 64) {
    return { isValid: false, msg: "Password cannot exceed 64 characters." };
  }

  let hasUpper = false;
  let hasLower = false;
  let hasDigit = false;
  let hasSymbol = false;

  for (const ch of pw) {
    if (/[A-Z]/.test(ch)) {
      hasUpper = true;
    } else if (/[a-z]/.test(ch)) {
      hasLower = true;
    } else if (/[0-9]/.test(ch)) {
      hasDigit = true;
    } else if (isAllowedSymbol(ch)) {
      hasSymbol = true;
    }
  }

  if (!hasUpper) {
    return { isValid: false, msg: "Password must contain at least one uppercase letter." };
  }
  if (!hasLower) {
    return { isValid: false, msg: "Password must contain at least one lowercase letter." };
  }
  if (!hasDigit) {
    return { isValid: false, msg: "Password must contain at least one number." };
  }
  if (!hasSymbol) {
    return { isValid: false, msg: "Password must contain at least one symbol (!@#$%^&*)." };
  }

  return { isValid: true, msg: "" };
}

function isAllowedSymbol(ch: string): boolean {
  const symbols = "!@#$%^&*";
  return symbols.includes(ch);
}