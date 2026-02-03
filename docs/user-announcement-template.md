# User Communication Template

**Note**: This is a template for future user announcement. Deployment and communication timing is controlled by the project owner.

---

## Email Announcement

**Subject**: Improved Authentication - Longer Sessions, Better Experience

**Body**:

Hello PimpMyPack Community,

We've improved our authentication system to make your experience smoother and more secure.

**What's New:**
- Automatic session extension while you're actively using the app
- Optional "Remember Me" for up to 30 days on trusted devices
- Enhanced security with shorter access tokens + secure refresh tokens

**What This Means for You:**
- No action required - works automatically
- Fewer unexpected logouts
- Better mobile/SPA experience
- Optional 30-day "remember me"

**For Developers:**
See [docs/frontend-integration.md](frontend-integration.md) for implementation details.

**Questions?**
Contact support@pimpmypack.example.com

The PimpMyPack Team

---

## In-App Notification (Optional)

**Title**: 🔐 New: Improved Authentication

**Message**:
Your sessions now last longer! We've added automatic token refresh so you won't get logged out unexpectedly. Enable "Remember Me" at login to stay logged in for 30 days.

[Learn More](#)

---

## Social Media Post (Optional)

🔐 Just launched improved authentication for PimpMyPack!

✅ Automatic session extension
✅ Optional 30-day "remember me"
✅ Better security + smoother UX

For developers: Check out our comprehensive frontend integration guide 📚

#authentication #security #devtools

---

## Developer Announcement (GitHub/Slack/Discord)

**Title**: 🎉 Refresh Token Authentication Now Available

Hey developers!

We've just released refresh token authentication for PimpMyPack. This brings automatic session extension and better UX for your users.

**Key Features:**
- POST `/auth/refresh` endpoint for token refresh
- "Remember me" option (up to 30 days)
- Rate limiting (10 req/min per IP)
- Comprehensive audit logging
- Backward compatible (existing clients continue to work)

**Getting Started:**
📖 Read the [Frontend Integration Guide](frontend-integration.md) for:
- Automatic refresh implementation
- Storage best practices
- React hooks & Vue composables
- Error handling patterns

**Configuration:**
New environment variables available:
- `REFRESH_TOKEN_DAYS` (default: 1 day)
- `REFRESH_TOKEN_REMEMBER_ME_DAYS` (default: 30 days)
- `REFRESH_RATE_LIMIT_REQUESTS` (default: 10)

See `.env.sample` for complete configuration options.

**Migration:**
Database migration `000006_refresh_tokens` runs automatically on app start. No manual intervention required.

Happy coding! 🚀

---

## FAQ Template (Optional)

**Q: Do I need to update my frontend code?**
A: No, existing clients continue to work. The refresh token feature is opt-in.

**Q: What happens to my current session?**
A: Nothing changes. Your current login session remains valid.

**Q: How do I enable "Remember Me"?**
A: Check the "Remember Me" box when logging in, or pass `"remember_me": true` in your login API request.

**Q: Is this more secure?**
A: Yes! Access tokens now expire after 15 minutes (down from 60), reducing the window of exposure. Refresh tokens enable automatic session extension without compromising security.

**Q: What if I encounter issues?**
A: Contact support at support@pimpmypack.example.com or open an issue on GitHub.

---

## Release Notes Section

See [CHANGELOG.md](../CHANGELOG.md) for complete technical details and developer notes.
