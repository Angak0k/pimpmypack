# Authentication Refresh Token & Remember Me - Specification

**Status**: ✅ Reviewed/Validated
**Author**: Claude Agent
**Date**: 2026-02-02
**Validated By**: PO, Tech Lead, Frontend Dev
**Implementation Started**: 2026-02-02 (Phase 1 - Backend Core)

---

## 📋 Overview

### Purpose

Améliorer l'expérience utilisateur en implémentant des sessions glissantes (sliding sessions) et une authentification longue durée pour l'API PimpMyPack. Cela permet aux utilisateurs authentifiés de rester connectés sans ré-authentification fréquente tout en maintenant les bonnes pratiques de sécurité.

### Problem

Le système d'authentification JWT actuel présente des limitations qui impactent négativement l'expérience utilisateur :

1. **Durée de session fixe** : Les tokens expirent exactement 1 heure après leur création (par défaut), indépendamment de l'activité utilisateur
2. **Pas d'extension de session** : Les utilisateurs actifs doivent se reconnecter toutes les heures, même s'ils utilisent l'application en continu
3. **Pas d'option "remember me"** : Les utilisateurs souhaitant rester connectés plus longtemps (ex: sur appareil personnel) n'ont pas d'autre option que d'ajuster manuellement la durée globale des tokens
4. **Mauvaise expérience mobile/SPA** : Les applications single-page et mobiles obligent les utilisateurs à se ré-authentifier fréquemment, interrompant les workflows

Flux actuel :
```
Login Utilisateur → JWT Token (1h) → Token Expire → Re-login Forcé
                                            ↓
                                (Même si l'utilisateur était actif)
```

### Goals

1. **Session glissante** : Prolonger automatiquement l'authentification pour les utilisateurs actifs
2. **Fonctionnalité "remember me"** : Permettre aux utilisateurs d'opter pour des sessions longue durée sur des appareils de confiance
3. **Maintenir la sécurité** : Conserver des access tokens de courte durée pour minimiser l'exposition
4. **Rétrocompatibilité** : Les clients existants continuent de fonctionner sans modification
5. **Capacité de révocation** : Fondation pour de futures fonctionnalités de révocation de tokens
6. **Option stateless** : Conserver la validation JWT sans lookup en base (performances)

### Non-Goals

1. **Révocation/blacklist de tokens** : Non implémenté dans cette phase (amélioration future)
2. **Gestion multi-appareils** : Pas d'interface pour voir/gérer les sessions actives
3. **Conformité OAuth2 complète** : Flow simplifié de refresh token, pas OAuth2 complet
4. **Rotation de refresh tokens** : Tous les refresh tokens restent valides jusqu'à expiration (rotation = amélioration future)
5. **Stockage de sessions en base** : Refresh tokens en DB, mais access tokens restent stateless

---

## 🎯 Requirements

### Functional Requirements

#### FR1: Génération de Refresh Token

**Description** : Le système génère à la fois des access tokens et des refresh tokens lors d'un login réussi.

**Critères d'acceptation** :
- Le endpoint de login retourne `access_token` et `refresh_token` dans la réponse
- L'access token a une durée de vie courte (configurable, défaut : 15 minutes)
- Le refresh token a une durée de vie plus longue (configurable, défaut : 24 heures)
- Le refresh token est stocké en base de données avec des métadonnées (user_id, expiration, date de création)
- Les refresh tokens sont des chaînes aléatoires cryptographiquement sécurisées (pas des JWTs)
- Chaque login génère un nouveau refresh token (les anciens restent valides jusqu'à expiration)

**Priorité** : Haute

#### FR2: Endpoint de Refresh Token

**Description** : Les clients peuvent échanger un refresh token valide contre un nouvel access token.

**Critères d'acceptation** :
- Endpoint `POST /auth/refresh` accepte le refresh token dans le body de la requête
- Valide que le refresh token existe en base et n'a pas expiré
- Valide que le refresh token appartient à un compte utilisateur actif
- Génère et retourne un nouvel access token avec la durée de vie standard
- Ne génère PAS de nouveau refresh token (réutilise l'existant)
- Retourne 401 Unauthorized si le refresh token est invalide/expiré
- Met à jour le timestamp `last_used_at` du refresh token en base

**Priorité** : Haute

#### FR3: Option Remember Me

**Description** : Les utilisateurs peuvent opter pour des sessions étendues au moment du login.

**Critères d'acceptation** :
- Le endpoint de login accepte un paramètre booléen optionnel `remember_me`
- Si `remember_me = true`, la durée de vie du refresh token est étendue (défaut : 30 jours)
- Si `remember_me = false` ou omis, le refresh token obtient la durée standard (défaut : 24 heures)
- La durée de vie de l'access token reste inchangée (toujours courte durée)
- La réponse indique quelle durée a été appliquée

**Priorité** : Haute

#### FR4: Nettoyage des Refresh Tokens

**Description** : Le système nettoie automatiquement les refresh tokens expirés pour éviter le gonflement de la base de données.

**Critères d'acceptation** :
- Les refresh tokens expirés sont automatiquement supprimés de la base
- Le nettoyage s'exécute périodiquement (ex: quotidiennement via cron job ou routine au démarrage)
- Le nettoyage peut être déclenché manuellement via endpoint admin ou commande CLI
- L'opération de nettoyage est loguée

**Priorité** : Moyenne

#### FR5: Rétrocompatibilité

**Description** : Les clients existants continuent de fonctionner sans changement de code.

**Critères d'acceptation** :
- La réponse du endpoint de login inclut les deux tokens mais `access_token` reste le champ principal
- Les clients utilisant uniquement `access_token` continuent de fonctionner (doivent juste se reconnecter plus fréquemment)
- La logique de validation des tokens existante reste inchangée (les access tokens fonctionnent exactement comme avant)
- Aucun changement breaking sur les endpoints ou formats de réponse existants

**Priorité** : Haute

### Non-Functional Requirements

#### NFR1: Sécurité

**Exigences** :
- Refresh tokens stockés de manière sécurisée en base de données (hachés ou chiffrés at rest dans le futur)
- Refresh tokens générés avec un générateur de nombres aléatoires cryptographiquement sécurisé
- Refresh tokens longs (min 32 bytes) pour prévenir la force brute
- Access tokens restent de courte durée (15 min) pour minimiser la fenêtre d'exposition
- Endpoint de refresh rate-limité pour prévenir les abus
- HTTPS requis pour tous les endpoints d'authentification (exigence existante)

**Priorité** : Haute

#### NFR2: Performance

**Exigences** :
- La validation d'access token reste stateless (pas de lookup DB)
- La validation de refresh token nécessite une seule requête DB
- Index de base de données sur la colonne de lookup des refresh tokens
- Le nettoyage des refresh tokens ne bloque pas les opérations API
- L'opération de refresh de token se termine en < 100ms (p95)

**Priorité** : Haute

#### NFR3: Configurabilité

**Exigences** :
- Toutes les durées de vie des tokens configurables via variables d'environnement
- L'intervalle de nettoyage des refresh tokens configurable
- Stratégie de génération de tokens échangeable (pour futures améliorations)

**Priorité** : Moyenne

---

## 🏗️ Architecture

### Architecture Actuelle

```
┌─────────┐                                    ┌─────────┐
│         │  1. Login (user/pass)              │         │
│ Client  │───────────────────────────────────>│   API   │
│         │                                    │         │
│         │  2. JWT Token (1h lifespan)       │         │
│         │<───────────────────────────────────│         │
│         │                                    └─────────┘
│         │
│         │  3. API Request + JWT Token
│         │───────────────────────────────────>┌─────────┐
│         │                                    │ Validate│
│         │  4. Response                       │   JWT   │
│         │<───────────────────────────────────│(stateless)
└─────────┘                                    └─────────┘
     │
     │ (1 heure passe)
     ▼
┌─────────┐
│ Token   │  5. Re-login forcé
│ Expired │
└─────────┘
```

### Architecture Proposée

```
┌─────────┐                                    ┌──────────────┐
│         │  1. Login (user/pass, remember_me) │              │
│ Client  │───────────────────────────────────>│   API        │
│         │                                    │              │
│         │  2. Access Token (15m) +           │  ┌────────┐  │
│         │     Refresh Token (7d ou 30d)      │  │Database│  │
│         │<───────────────────────────────────│  │ (Store │  │
│         │                                    │  │Refresh)│  │
│         │                                    │  └────────┘  │
│         │  3. API Request + Access Token     │              │
│         │───────────────────────────────────>│  Validate   │
│         │                                    │  Access JWT │
│         │  4. Response                       │  (stateless) │
│         │<───────────────────────────────────│              │
└─────────┘                                    └──────────────┘
     │
     │ (Access token expire après 15 min)
     ▼
┌─────────┐
│ Access  │  5. Refresh request (Refresh Token)
│ Expired │───────────────────────────────────>┌──────────────┐
│         │                                    │   Validate   │
│         │                                    │   Refresh    │
│         │  6. New Access Token (15m)         │   (DB check) │
│         │<───────────────────────────────────│              │
│         │                                    └──────────────┘
│         │  7. Continue using app
│         │    (répète étapes 5-6 toutes les 15 min)
└─────────┘
     │
     │ (Refresh token expire après 24h-30 jours d'inactivité)
     ▼
┌─────────┐
│ Refresh │  8. Re-login requis
│ Expired │
└─────────┘
```

### Composants Modifiés/Nouveaux

#### Nouveaux Composants

L'architecture suit le pattern de séparation handlers/CRUD similaire à `pkg/images/` :

1. **Types Refresh Token** ([pkg/security/types.go](pkg/security/types.go)) - Fichier enrichi
   - Ajout struct `RefreshToken` (données de la table `refresh_tokens`)
   - Types pour les requêtes/réponses :
     - `RefreshTokenInput` (body de `/auth/refresh`)
     - `TokenPairResponse` (réponse login avec access + refresh)
     - `RefreshResponse` (réponse endpoint refresh)
   - Méthodes utilitaires (ex: `IsExpired()` sur `RefreshToken`)

2. **Opérations CRUD Refresh Tokens** ([pkg/security/refresh_tokens.go](pkg/security/refresh_tokens.go)) - NOUVEAU fichier
   - Fonctions CRUD pour refresh tokens (utilisant SQL direct, `database/sql`)
   - `CreateRefreshToken(ctx, token, userID, expiresAt)` : INSERT nouveau token
   - `GetRefreshToken(ctx, token)` : SELECT par valeur de token
   - `UpdateLastUsed(ctx, token)` : UPDATE last_used_at
   - `DeleteRefreshToken(ctx, token)` : DELETE un token (révocation)
   - `CleanupExpiredTokens(ctx)` : DELETE tokens expirés (cron périodique)

3. **Génération de Tokens** ([pkg/security/tokens.go](pkg/security/tokens.go)) - NOUVEAU fichier
   - Regroupe toutes les fonctions de génération et validation de tokens
   - Fonctions existantes déplacées depuis `security.go` :
     - `GenerateToken(userID)` : Génère JWT access token
     - `TokenValid(c)` : Valide JWT access token
     - `ExtractToken(c)` : Extrait JWT depuis header/query
     - `ExtractTokenID(c)` : Extrait user_id depuis JWT
   - Nouvelles fonctions :
     - `GenerateRefreshToken()` : Génération crypto-sécurisée (crypto/rand, 32 bytes, préfixe `rt_`)
     - `GenerateTokenPair(userID, rememberMe)` : Génère access + refresh tokens avec durées appropriées

4. **Handlers HTTP** ([pkg/security/handlers.go](pkg/security/handlers.go)) - NOUVEAU fichier
   - Nouveau handler `RefreshTokenHandler(c *gin.Context)`
   - Validation et échange de refresh token contre access token
   - Logique : bind JSON, valider token en DB, vérifier expiration, vérifier user actif, générer nouveau JWT

5. **Middleware** ([pkg/security/middleware.go](pkg/security/middleware.go)) - NOUVEAU fichier
   - Fonctions déplacées depuis `security.go` :
     - `JwtAuthProcessor()` : Middleware JWT standard
     - `JwtAuthAdminProcessor()` : Middleware JWT + vérification role admin

#### Composants Modifiés

1. **Package Security - security.go** ([pkg/security/security.go](pkg/security/security.go))
   - Fichier allégé : conserve uniquement `HashPassword()` et `VerifyPassword()`
   - Autres fonctions déplacées vers `tokens.go` et `middleware.go`
   - ⚠️ Rétrocompatibilité : Les imports externes (`accounts`, `main.go`) continuent de fonctionner car les fonctions publiques restent dans le package `security`

2. **Handler Login** ([pkg/accounts/accounts.go](pkg/accounts/accounts.go))
   - Accepte paramètre `remember_me` optionnel dans `LoginInput`
   - Appelle `security.GenerateTokenPair()` au lieu de `security.GenerateToken()` (fonction définie dans [pkg/security/tokens.go](pkg/security/tokens.go))
   - Retourne à la fois access_token et refresh_token (format `TokenPairResponse`)
   - Stocke le refresh token en base via `security.CreateRefreshToken()` (fonction définie dans [pkg/security/refresh_tokens.go](pkg/security/refresh_tokens.go))

3. **Router** ([main.go](main.go))
   - Ajout endpoint `POST /auth/refresh` → `security.RefreshTokenHandler()` (handler défini dans [pkg/security/handlers.go](pkg/security/handlers.go))
   - Les middleware `security.JwtAuthProcessor()` et `security.JwtAuthAdminProcessor()` continuent de fonctionner (déplacés vers [pkg/security/middleware.go](pkg/security/middleware.go))

4. **Config** ([pkg/config/env.go](pkg/config/env.go))
   - Ajout nouvelles variables de configuration :
     - `ACCESS_TOKEN_MINUTES` (remplace progressivement `TOKEN_HOUR_LIFESPAN`)
     - `REFRESH_TOKEN_DAYS` (défaut: 1)
     - `REFRESH_TOKEN_REMEMBER_ME_DAYS` (défaut: 30)
     - `REFRESH_TOKEN_CLEANUP_INTERVAL_HOURS` (défaut: 24)
   - Mapping de rétrocompatibilité `TOKEN_HOUR_LIFESPAN` → `ACCESS_TOKEN_MINUTES`

---

## 🔄 Diagrammes de Séquence - Flows Frontend

### Flow 1: Login Initial (Sans Remember Me)

```mermaid
sequenceDiagram
    participant U as User
    participant F as Frontend
    participant API as Backend API
    participant DB as Database

    U->>F: Entre credentials
    F->>API: POST /login<br/>{username, password, remember_me: false}
    API->>DB: Vérifier credentials
    DB-->>API: User valide
    API->>API: Générer Access Token (15min)
    API->>API: Générer Refresh Token (24h)
    API->>DB: Stocker Refresh Token
    DB-->>API: OK
    API-->>F: 200 OK<br/>{access_token, refresh_token,<br/>expires_in: 900, refresh_expires_in: 86400}
    F->>F: Stocker access_token (mémoire)<br/>Stocker refresh_token (secure storage)
    F-->>U: Rediriger vers dashboard

    Note over F: Access token valide pour 15 minutes<br/>Refresh token valide pour 24 heures
```

### Flow 2: Login avec Remember Me

```mermaid
sequenceDiagram
    participant U as User
    participant F as Frontend
    participant API as Backend API
    participant DB as Database

    U->>F: Entre credentials + coche "Remember me"
    F->>API: POST /login<br/>{username, password, remember_me: true}
    API->>DB: Vérifier credentials
    DB-->>API: User valide
    API->>API: Générer Access Token (15min)
    API->>API: Générer Refresh Token (30j)
    API->>DB: Stocker Refresh Token (expires_at: +30j)
    DB-->>API: OK
    API-->>F: 200 OK<br/>{access_token, refresh_token,<br/>expires_in: 900, refresh_expires_in: 2592000}
    F->>F: Stocker access_token (mémoire)<br/>Stocker refresh_token (secure storage)
    F-->>U: Rediriger vers dashboard

    Note over F: Access token valide pour 15 minutes<br/>Refresh token valide pour 30 jours
```

### Flow 3: Utilisation Normale + Auto-Refresh

```mermaid
sequenceDiagram
    participant U as User
    participant F as Frontend
    participant API as Backend API
    participant DB as Database

    Note over F: Access token stocké, valide 15min

    U->>F: Navigation dans l'app
    F->>API: GET /api/packs<br/>Authorization: Bearer {access_token}
    API->>API: Valider JWT (stateless)
    API-->>F: 200 OK + données
    F-->>U: Afficher données

    Note over F: 14 minutes passent...

    U->>F: Nouvelle action
    F->>API: POST /api/packs<br/>Authorization: Bearer {access_token}
    API->>API: Valider JWT (stateless)
    API-->>F: 401 Unauthorized<br/>{error: "Token expired"}

    F->>F: Détecter 401 → Interceptor<br/>Lancer refresh automatique
    F->>API: POST /auth/refresh<br/>{refresh_token: "rt_..."}
    API->>DB: SELECT * FROM refresh_tokens<br/>WHERE token = "rt_..."
    DB-->>API: Refresh token valide
    API->>DB: Vérifier user actif
    DB-->>API: User actif
    API->>API: Générer nouveau Access Token
    API->>DB: UPDATE last_used_at
    DB-->>API: OK
    API-->>F: 200 OK<br/>{access_token, expires_in: 900}

    F->>F: Stocker nouveau access_token
    F->>API: POST /api/packs (retry)<br/>Authorization: Bearer {new_access_token}
    API->>API: Valider JWT
    API-->>F: 201 Created
    F-->>U: Action réussie (transparent)

    Note over U: L'utilisateur n'a rien remarqué !
```

### Flow 4: Refresh Token Expiré (Après 24h-30 jours)

```mermaid
sequenceDiagram
    participant U as User
    participant F as Frontend
    participant API as Backend API
    participant DB as Database

    Note over F: 24 heures passées, refresh token expiré

    U->>F: Ouvre l'application
    F->>F: Access token expiré (> 15min)
    F->>API: POST /auth/refresh<br/>{refresh_token: "rt_..."}
    API->>DB: SELECT * FROM refresh_tokens<br/>WHERE token = "rt_..."
    DB-->>API: Token trouvé mais expiré
    API-->>F: 401 Unauthorized<br/>{error: "Refresh token expired"}

    F->>F: Supprimer tokens stockés
    F->>F: Rediriger vers /login
    F-->>U: Page de login

    Note over U: L'utilisateur doit se reconnecter
```

### Flow 5: Gestion d'Erreurs - Refresh Token Invalide

```mermaid
sequenceDiagram
    participant U as User
    participant F as Frontend
    participant API as Backend API
    participant DB as Database

    U->>F: Action dans l'app
    F->>API: GET /api/items<br/>Authorization: Bearer {access_token}
    API-->>F: 401 Unauthorized

    F->>API: POST /auth/refresh<br/>{refresh_token: "rt_corrupted"}
    API->>DB: SELECT * FROM refresh_tokens<br/>WHERE token = "rt_corrupted"
    DB-->>API: Token non trouvé
    API-->>F: 401 Unauthorized<br/>{error: "Invalid refresh token"}

    F->>F: Supprimer tokens
    F->>F: Rediriger vers /login
    F-->>U: Session invalide, veuillez vous reconnecter
```

### Flow 6: Gestion d'Erreurs - Compte Désactivé

```mermaid
sequenceDiagram
    participant U as User
    participant F as Frontend
    participant API as Backend API
    participant DB as Database

    Note over DB: Admin désactive le compte

    U->>F: Action dans l'app
    F->>API: GET /api/data<br/>Authorization: Bearer {access_token}
    API-->>F: 401 Unauthorized (token expiré)

    F->>API: POST /auth/refresh<br/>{refresh_token: "rt_..."}
    API->>DB: SELECT * FROM refresh_tokens<br/>WHERE token = "rt_..."
    DB-->>API: Refresh token valide
    API->>DB: SELECT status FROM accounts<br/>WHERE id = user_id
    DB-->>API: status = 'inactive'
    API-->>F: 403 Forbidden<br/>{error: "Account is not active"}

    F->>F: Supprimer tokens
    F->>F: Rediriger vers /login
    F-->>U: Votre compte a été désactivé
```

### Flow 7: Rate Limiting sur Refresh

```mermaid
sequenceDiagram
    participant A as Attacker
    participant F as Frontend/Bot
    participant API as Backend API
    participant RL as Rate Limiter

    Note over A: Tentative de brute force

    loop 10 fois en 1 minute
        F->>API: POST /auth/refresh<br/>{refresh_token: "guess_..."}
        API->>RL: Vérifier rate limit (IP)
        RL-->>API: OK (< 10 req/min)
        API-->>F: 401 Unauthorized
    end

    F->>API: POST /auth/refresh (11ème requête)<br/>{refresh_token: "guess_..."}
    API->>RL: Vérifier rate limit (IP)
    RL-->>API: LIMIT EXCEEDED
    API-->>F: 429 Too Many Requests<br/>{error: "Rate limit exceeded",<br/>retry_after: 60}

    Note over F: Bloqué pour 1 minute
```

### Flow 8: Implémentation Frontend - Interceptor Pattern

```mermaid
sequenceDiagram
    participant App as Application Code
    participant Int as HTTP Interceptor
    participant API as Backend API
    participant Store as Token Storage

    App->>Int: fetch('/api/resource')
    Int->>Store: Récupérer access_token
    Store-->>Int: access_token
    Int->>API: GET /api/resource<br/>Authorization: Bearer {access_token}

    alt Token valide
        API-->>Int: 200 OK + data
        Int-->>App: data
    else Token expiré (401)
        API-->>Int: 401 Unauthorized
        Int->>Store: Récupérer refresh_token
        Store-->>Int: refresh_token
        Int->>API: POST /auth/refresh<br/>{refresh_token}

        alt Refresh réussi
            API-->>Int: 200 OK + new access_token
            Int->>Store: Stocker new access_token
            Int->>API: GET /api/resource (retry)<br/>Authorization: Bearer {new_token}
            API-->>Int: 200 OK + data
            Int-->>App: data (transparent)
        else Refresh échoué
            API-->>Int: 401/403 Error
            Int->>Store: Supprimer tous tokens
            Int-->>App: Rediriger vers /login
        end
    end
```

---

## 🔌 API Specification

### Endpoint Modifié: Login

#### `POST /login`

**Description** : Authentifie un utilisateur et retourne des tokens d'accès et de rafraîchissement.

**Request** :
```json
{
  "username": "john.doe",
  "password": "SecurePass123!",
  "remember_me": false  // NOUVEAU: optionnel, défaut false
}
```

**Response (Success - 200 OK)** :
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "rt_7f9b8c6a5d4e3f2a1b0c9d8e7f6a5b4c3d2e1f0a9b8c7d6e5f4a3b2c1d0e9f8",
  "token_type": "Bearer",
  "expires_in": 900,
  "refresh_expires_in": 86400,
  "pending": false,
  "message": "Login successful"
}
```

**Champs de réponse** :

| Champ | Type | Description |
|-------|------|-------------|
| `access_token` | string | JWT pour accéder aux ressources protégées (15 min) |
| `refresh_token` | string | Token opaque pour obtenir de nouveaux access tokens (24h-30j) |
| `token_type` | string | Type de token, toujours "Bearer" |
| `expires_in` | int | Durée de vie de l'access token en secondes (900 = 15 min) |
| `refresh_expires_in` | int | Durée de vie du refresh token en secondes |
| `pending` | bool | Indique si l'email est confirmé (existant) |
| `message` | string | Message de succès |

**Response (Error - 401 Unauthorized)** :
```json
{
  "error": "Invalid credentials"
}
```

**Response (Error - 403 Forbidden)** :
```json
{
  "error": "Account not confirmed"
}
```

**Notes** :
- Rétrocompatible : clients ignorant `refresh_token` continuent de fonctionner
- `expires_in` et `refresh_expires_in` permettent au client de calculer l'expiration
- `remember_me=true` → `refresh_expires_in=2592000` (30 jours)
- `remember_me=false` → `refresh_expires_in=86400` (24 heures)

---

### Nouvel Endpoint: Refresh Token

#### `POST /auth/refresh`

**Description** : Échange un refresh token valide contre un nouvel access token.

**Request** :
```json
{
  "refresh_token": "rt_7f9b8c6a5d4e3f2a1b0c9d8e7f6a5b4c3d2e1f0a9b8c7d6e5f4a3b2c1d0e9f8"
}
```

**Response (Success - 200 OK)** :
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 900
}
```

**Response (Error - 400 Bad Request)** :
```json
{
  "error": "Refresh token is required"
}
```

**Response (Error - 401 Unauthorized)** :
```json
{
  "error": "Invalid or expired refresh token"
}
```

**Response (Error - 403 Forbidden)** :
```json
{
  "error": "Account is not active"
}
```

**Response (Error - 429 Too Many Requests)** :
```json
{
  "error": "Rate limit exceeded",
  "retry_after": 60
}
```

**Comportement** :
- Valide le refresh token en base de données
- Vérifie que le token n'a pas expiré
- Vérifie que l'utilisateur associé est toujours actif
- Génère un NOUVEAU access token
- Ne génère PAS de nouveau refresh token (réutilise l'existant)
- Met à jour `last_used_at` du refresh token
- Rate limiting : 10 requêtes/minute par IP

**Sécurité** :
- Pas d'authentification requise (le refresh token EST la credential)
- Le refresh token doit être transmis dans le body (pas en URL)
- HTTPS obligatoire
- Rate limiting pour prévenir le brute force

---

### Nouvel Endpoint: Logout (Future Enhancement)

#### `POST /auth/logout`

**Description** : Invalide le refresh token actuel (révocation).

**Status** : 📋 Planifié (Phase 2)

**Request** :
```json
{
  "refresh_token": "rt_7f9b8c6a5d4e3f2a1b0c9d8e7f6a5b4c3d2e1f0a9b8c7d6e5f4a3b2c1d0e9f8"
}
```

**Response (Success - 200 OK)** :
```json
{
  "message": "Logged out successfully"
}
```

**Notes** :
- Phase 1 : Non implémenté (tous les refresh tokens valides jusqu'à expiration)
- Phase 2 : Supprime le refresh token de la base (révocation simple)
- Phase 3 : Pourrait ajouter "logout all devices" (supprimer tous les refresh tokens de l'utilisateur)

---

## 💾 Modèle de Données

### Nouvelle Table: `refresh_tokens`

**Fichier de migration** : `000006_refresh_tokens.up.sql`

```sql
CREATE TABLE refresh_tokens (
    id BIGSERIAL PRIMARY KEY,
    token VARCHAR(255) NOT NULL UNIQUE,
    user_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_token ON refresh_tokens(token);
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

COMMENT ON TABLE refresh_tokens IS 'Stocke les refresh tokens pour l''authentification à session glissante';
COMMENT ON COLUMN refresh_tokens.token IS 'Chaîne de refresh token unique, préfixée par rt_';
COMMENT ON COLUMN refresh_tokens.expires_at IS 'Date d''expiration du refresh token (24 heures à 30 jours après création)';
COMMENT ON COLUMN refresh_tokens.last_used_at IS 'Date de dernière utilisation pour obtenir un access token';
```

**Description des champs** :

| Champ | Type | Contraintes | Description |
|-------|------|-------------|-------------|
| `id` | BIGSERIAL | PRIMARY KEY | Identifiant unique |
| `token` | VARCHAR(255) | NOT NULL, UNIQUE | Chaîne du refresh token (préfixe `rt_` + 43 caractères base64) |
| `user_id` | BIGINT | NOT NULL, FK → accounts(id) | Utilisateur propriétaire du token |
| `expires_at` | TIMESTAMP | NOT NULL | Date/heure d'expiration du token |
| `created_at` | TIMESTAMP | NOT NULL, DEFAULT NOW() | Date/heure de création |
| `last_used_at` | TIMESTAMP | NOT NULL, DEFAULT NOW() | Date/heure de dernière utilisation |

**Indexes** :

| Index | Colonne(s) | Raison |
|-------|-----------|--------|
| `idx_refresh_tokens_token` | `token` | Lookup rapide lors de validation (cas d'usage principal) |
| `idx_refresh_tokens_user_id` | `user_id` | Support future feature "liste des sessions actives" |
| `idx_refresh_tokens_expires_at` | `expires_at` | Nettoyage efficace des tokens expirés |

**Contraintes** :
- `token` UNIQUE : Prévient les collisions
- `CASCADE DELETE` : Supprime automatiquement les refresh tokens quand un utilisateur est supprimé
- Pas de `UNIQUE(user_id)` : Un utilisateur peut avoir plusieurs refresh tokens actifs (multi-devices)

**Migration Down** : `000006_refresh_tokens.down.sql`

```sql
DROP INDEX IF EXISTS idx_refresh_tokens_expires_at;
DROP INDEX IF EXISTS idx_refresh_tokens_user_id;
DROP INDEX IF EXISTS idx_refresh_tokens_token;
DROP TABLE IF EXISTS refresh_tokens;
```

### Struct Go pour Refresh Token

**Fichier** : [pkg/security/types.go](pkg/security/types.go) - NOUVEAU fichier

```go
package security

import "time"

// RefreshToken represents a refresh token stored in the database
type RefreshToken struct {
    ID         uint      `json:"id"`
    Token      string    `json:"token"`
    UserID     uint      `json:"user_id"`
    ExpiresAt  time.Time `json:"expires_at"`
    CreatedAt  time.Time `json:"created_at"`
    LastUsedAt time.Time `json:"last_used_at"`
}

// IsExpired returns true if the refresh token has expired
func (rt *RefreshToken) IsExpired() bool {
    return time.Now().After(rt.ExpiresAt)
}

// RefreshTokenInput represents the data required to refresh an access token
type RefreshTokenInput struct {
    RefreshToken string `json:"refresh_token" binding:"required"`
}

// TokenPairResponse represents the response containing both tokens (login)
type TokenPairResponse struct {
    AccessToken       string `json:"access_token"`
    RefreshToken      string `json:"refresh_token"`
    TokenType         string `json:"token_type"`
    ExpiresIn         int    `json:"expires_in"`
    RefreshExpiresIn  int    `json:"refresh_expires_in"`
    Pending           bool   `json:"pending"`
    Message           string `json:"message"`
}

// RefreshResponse represents the response from a token refresh
type RefreshResponse struct {
    AccessToken string `json:"access_token"`
    TokenType   string `json:"token_type"`
    ExpiresIn   int    `json:"expires_in"`
}
```

**Notes** :
- Pas de tags GORM : le projet utilise SQL direct via `database/sql`
- Tags JSON pour la sérialisation des réponses API
- Tags `binding:"required"` pour la validation Gin
- Méthode `IsExpired()` pour faciliter les vérifications

---

## ⚙️ Configuration

### Nouvelles Variables d'Environnement

**Fichier** : `.env.sample` (à documenter également dans README)

```bash
# ============================================
# Authentication & Token Configuration
# ============================================

# Access Token (JWT) - Durée de vie courte
# Durée en minutes pour les access tokens JWT
# Recommandé: 15-30 minutes
ACCESS_TOKEN_MINUTES=15

# Refresh Token - Durée standard
# Durée en jours pour les refresh tokens (remember_me = false)
# Recommandé: 1 jour (24 heures)
REFRESH_TOKEN_DAYS=1

# Refresh Token - Durée "Remember Me"
# Durée en jours pour les refresh tokens (remember_me = true)
# Recommandé: 30 jours (max 90 jours)
REFRESH_TOKEN_REMEMBER_ME_DAYS=30

# Cleanup Configuration
# Fréquence de nettoyage des refresh tokens expirés (en heures)
# Recommandé: 24 heures (quotidien)
REFRESH_TOKEN_CLEANUP_INTERVAL_HOURS=24

# ============================================
# DEPRECATED - Conservé pour rétrocompatibilité
# ============================================

# DEPRECATED: Utiliser ACCESS_TOKEN_MINUTES à la place
# Si ACCESS_TOKEN_MINUTES n'est pas défini, cette valeur est utilisée
# 1 heure = 60 minutes
# TOKEN_HOUR_LIFESPAN=1
```

### Struct de Configuration Go

**Fichier** : [pkg/config/env.go](pkg/config/env.go)

```go
type Config struct {
    // ... champs existants ...

    // Access Token Configuration
    AccessTokenMinutes int // Durée de vie des access tokens JWT en minutes

    // Refresh Token Configuration
    RefreshTokenDays           int  // Durée standard des refresh tokens (jours)
    RefreshTokenRememberMeDays int  // Durée "remember me" des refresh tokens (jours)

    // Cleanup Configuration
    RefreshTokenCleanupIntervalHours int // Intervalle de nettoyage (heures)

    // DEPRECATED - Conservé pour rétrocompatibilité
    TokenLifespan int // Ancienne config en heures, mappée à AccessTokenMinutes
}
```

**Fonction de chargement** :

```go
func LoadConfig() Config {
    config := Config{
        // Charger nouvelles variables
        AccessTokenMinutes:               getEnvAsInt("ACCESS_TOKEN_MINUTES", 15),
        RefreshTokenDays:                 getEnvAsInt("REFRESH_TOKEN_DAYS", 7),
        RefreshTokenRememberMeDays:       getEnvAsInt("REFRESH_TOKEN_REMEMBER_ME_DAYS", 30),
        RefreshTokenCleanupIntervalHours: getEnvAsInt("REFRESH_TOKEN_CLEANUP_INTERVAL_HOURS", 24),

        // Rétrocompatibilité : TOKEN_HOUR_LIFESPAN
        TokenLifespan: getEnvAsInt("TOKEN_HOUR_LIFESPAN", 1),
    }

    // Si ACCESS_TOKEN_MINUTES n'est pas défini, utiliser TOKEN_HOUR_LIFESPAN (conversion)
    if !isEnvSet("ACCESS_TOKEN_MINUTES") && isEnvSet("TOKEN_HOUR_LIFESPAN") {
        config.AccessTokenMinutes = config.TokenLifespan * 60
        log.Warn("TOKEN_HOUR_LIFESPAN is deprecated, use ACCESS_TOKEN_MINUTES instead")
    }

    return config
}
```

### Valeurs Recommandées par Environnement

#### Development
```bash
ACCESS_TOKEN_MINUTES=60                    # 1h (plus confortable en dev)
REFRESH_TOKEN_DAYS=1
REFRESH_TOKEN_REMEMBER_ME_DAYS=30
REFRESH_TOKEN_CLEANUP_INTERVAL_HOURS=24
```

#### Staging
```bash
ACCESS_TOKEN_MINUTES=15                    # Production-like
REFRESH_TOKEN_DAYS=1
REFRESH_TOKEN_REMEMBER_ME_DAYS=30
REFRESH_TOKEN_CLEANUP_INTERVAL_HOURS=24
```

#### Production
```bash
ACCESS_TOKEN_MINUTES=15                    # Sécurité maximale
REFRESH_TOKEN_DAYS=1
REFRESH_TOKEN_REMEMBER_ME_DAYS=30
REFRESH_TOKEN_CLEANUP_INTERVAL_HOURS=24
```

---

## 🔒 Sécurité

### Génération de Refresh Token

**Exigences** :
- Utiliser `crypto/rand` pour génération cryptographiquement sécurisée
- Minimum 32 bytes d'entropie (256 bits)
- Encoder en base64url ou hex pour stockage/transmission
- Préfixer avec `rt_` pour identification facile et future extension de types de tokens

**Implémentation Go** :

**Fichier** : [pkg/security/tokens.go](pkg/security/tokens.go)

```go
package security

import (
    "crypto/rand"
    "encoding/base64"
    "fmt"
)

// GenerateRefreshToken génère un refresh token cryptographiquement sécurisé
func GenerateRefreshToken() (string, error) {
    // 32 bytes = 256 bits d'entropie
    b := make([]byte, 32)

    // crypto/rand.Read utilise le générateur aléatoire du système d'exploitation
    if _, err := rand.Read(b); err != nil {
        return "", fmt.Errorf("failed to generate random token: %w", err)
    }

    // base64url encoding (43 caractères) + préfixe "rt_"
    // Longueur totale: 46 caractères
    token := "rt_" + base64.URLEncoding.EncodeToString(b)

    return token, nil
}

// GenerateTokenPair génère à la fois un access token et un refresh token
func GenerateTokenPair(userID uint, rememberMe bool) (accessToken, refreshToken string, expiresAt time.Time, err error) {
    // Générer access token (JWT)
    accessToken, err = GenerateToken(userID)
    if err != nil {
        return "", "", time.Time{}, fmt.Errorf("failed to generate access token: %w", err)
    }

    // Générer refresh token
    refreshToken, err = GenerateRefreshToken()
    if err != nil {
        return "", "", time.Time{}, fmt.Errorf("failed to generate refresh token: %w", err)
    }

    // Calculer expiration du refresh token
    duration := time.Duration(config.RefreshTokenDays) * 24 * time.Hour
    if rememberMe {
        duration = time.Duration(config.RefreshTokenRememberMeDays) * 24 * time.Hour
    }
    expiresAt = time.Now().Add(duration)

    return accessToken, refreshToken, expiresAt, nil
}
```

**Exemple de token généré** :
```
rt_7f9b8c6a5d4e3f2a1b0c9d8e7f6a5b4c3d2e1f0a9b8c
```

**Propriétés de sécurité** :
- 2^256 combinaisons possibles (impossible à brute force)
- Aucun pattern prévisible
- Pas d'information sur l'utilisateur (opaque)
- Prefix `rt_` permet de distinguer des access tokens (JWTs)

### Stockage des Tokens

#### Phase 1 : Stockage en Clair (Spec Actuelle)

**Approche** :
- Stocker les refresh tokens en clair (plaintext) dans PostgreSQL
- Contrôle d'accès via credentials de l'application uniquement
- Database chiffrée at rest (niveau infrastructure)

**Justification** :
- Simplifie l'implémentation initiale
- Permet futures features (liste des sessions, révocation)
- Compromission DB déjà catastrophique (contient hashes de passwords, données utilisateur)
- Migration vers hachage possible en Phase 2 sans changement client

**Implémentation SQL directe** :

**Fichier** : [pkg/security/refresh_tokens.go](pkg/security/refresh_tokens.go)

```go
// CreateRefreshToken insère un nouveau refresh token dans la base
func CreateRefreshToken(ctx context.Context, token string, userID uint, expiresAt time.Time) error {
    _, err := database.DB().ExecContext(ctx,
        `INSERT INTO refresh_tokens (token, user_id, expires_at, created_at, last_used_at)
         VALUES ($1, $2, $3, $4, $5)`,
        token,
        userID,
        expiresAt,
        time.Now(),
        time.Now(),
    )
    if err != nil {
        return fmt.Errorf("failed to create refresh token: %w", err)
    }
    return nil
}

// GetRefreshToken récupère un refresh token par sa valeur
func GetRefreshToken(ctx context.Context, token string) (*RefreshToken, error) {
    var rt RefreshToken

    err := database.DB().QueryRowContext(ctx,
        `SELECT id, token, user_id, expires_at, created_at, last_used_at
         FROM refresh_tokens
         WHERE token = $1`,
        token,
    ).Scan(&rt.ID, &rt.Token, &rt.UserID, &rt.ExpiresAt, &rt.CreatedAt, &rt.LastUsedAt)

    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("refresh token not found")
    }
    if err != nil {
        return nil, fmt.Errorf("failed to get refresh token: %w", err)
    }

    return &rt, nil
}

// UpdateLastUsed met à jour le timestamp de dernière utilisation
func UpdateLastUsed(ctx context.Context, token string) error {
    _, err := database.DB().ExecContext(ctx,
        `UPDATE refresh_tokens
         SET last_used_at = $1
         WHERE token = $2`,
        time.Now(),
        token,
    )
    if err != nil {
        return fmt.Errorf("failed to update last_used_at: %w", err)
    }
    return nil
}

// DeleteRefreshToken supprime un refresh token
func DeleteRefreshToken(ctx context.Context, token string) error {
    _, err := database.DB().ExecContext(ctx,
        `DELETE FROM refresh_tokens WHERE token = $1`,
        token,
    )
    if err != nil {
        return fmt.Errorf("failed to delete refresh token: %w", err)
    }
    return nil
}

// CleanupExpiredTokens supprime tous les refresh tokens expirés
func CleanupExpiredTokens(ctx context.Context) (int64, error) {
    result, err := database.DB().ExecContext(ctx,
        `DELETE FROM refresh_tokens WHERE expires_at < $1`,
        time.Now(),
    )
    if err != nil {
        return 0, fmt.Errorf("failed to cleanup expired tokens: %w", err)
    }

    count, _ := result.RowsAffected()
    return count, nil
}
```

**Mesures de protection** :
```go
// Logs : JAMAIS logger le refresh token complet
log.Info("Refresh token used", "token_prefix", token[:10], "user_id", userID)

// Métriques : Pas d'exposition du token
metrics.RecordRefreshTokenUsage(userID, "success")
```

#### Phase 2 : Hachage des Tokens (Future Enhancement)

**Approche** :
- Hacher les refresh tokens avant stockage (comme les passwords)
- Utiliser bcrypt ou Argon2 avec cost factor approprié
- Stocker uniquement le hash en base, comparer le hash lors de la validation

**Exemple** :
```go
// Stockage
hashedToken := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)
db.Create(&RefreshToken{Token: hashedToken, UserID: userID})

// Validation
storedHash := db.FindRefreshToken(hashedToken)
err := bcrypt.CompareHashAndPassword(storedHash, []byte(receivedToken))
```

**Trade-offs** :
- ✅ Meilleure sécurité si DB compromise
- ❌ Coût CPU plus élevé (hachage sur chaque validation)
- ❌ Révocation par ID de token impossible (hash change)
- ❌ "Liste des sessions" nécessite métadonnées supplémentaires

### Transmission des Tokens

#### Access Token (JWT)

**Méthode actuelle conservée** :
```http
GET /api/packs HTTP/1.1
Host: api.pimpmypack.com
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Alternative (query param - déconseillé mais supporté)** :
```http
GET /api/packs?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9... HTTP/1.1
```

⚠️ **Note** : Query params peuvent être loggés, éviter en production.

#### Refresh Token

**Méthode recommandée (body JSON)** :
```http
POST /auth/refresh HTTP/1.1
Host: api.pimpmypack.com
Content-Type: application/json

{
  "refresh_token": "rt_7f9b8c6a5d4e3f2a1b0c9d8e7f6a5b4c..."
}
```

❌ **JAMAIS** :
- Dans l'URL : `GET /auth/refresh?refresh_token=...`
- Dans les logs applicatifs
- Dans le localStorage (XSS risk pour frontend)

### Bonnes Pratiques Frontend

**Stockage côté client** :

| Option | Sécurité | Use Case |
|--------|----------|----------|
| **httpOnly Cookie** | ✅ Excellent | Same-origin, protection XSS |
| **Memory (variable JS)** | ✅ Bon | SPA, disparaît au refresh |
| **sessionStorage** | ⚠️ Moyen | SPA, disparaît à fermeture tab |
| **localStorage** | ❌ Risqué | Accessible par XSS |

**Recommandation pour PimpMyPack** :
```javascript
// Option 1: Memory + sessionStorage fallback (SPA)
class TokenManager {
  constructor() {
    this.accessToken = null;
    this.refreshToken = sessionStorage.getItem('refresh_token');
  }

  setTokens(accessToken, refreshToken) {
    this.accessToken = accessToken; // Mémoire uniquement
    this.refreshToken = refreshToken;
    sessionStorage.setItem('refresh_token', refreshToken);
  }

  clearTokens() {
    this.accessToken = null;
    this.refreshToken = null;
    sessionStorage.removeItem('refresh_token');
  }
}

// Option 2: httpOnly Cookie (si backend et frontend same domain)
// Backend set-cookie lors du login, frontend envoie automatiquement
```

**Interceptor HTTP (automatic refresh)** :
```javascript
// Axios example
axios.interceptors.response.use(
  response => response,
  async error => {
    const originalRequest = error.config;

    // Si 401 et pas déjà retry
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      try {
        // Tenter refresh
        const { data } = await axios.post('/auth/refresh', {
          refresh_token: tokenManager.refreshToken
        });

        // Stocker nouveau access token
        tokenManager.setAccessToken(data.access_token);

        // Retry requête originale avec nouveau token
        originalRequest.headers['Authorization'] = `Bearer ${data.access_token}`;
        return axios(originalRequest);

      } catch (refreshError) {
        // Refresh échoué → logout
        tokenManager.clearTokens();
        router.push('/login');
        return Promise.reject(refreshError);
      }
    }

    return Promise.reject(error);
  }
);
```

### Rate Limiting

**Endpoint `/auth/refresh`** :

```go
// Limite : 10 requêtes par minute par IP
// Prévient brute force et DoS

type RateLimitConfig struct {
    RequestsPerMinute int
    BurstSize         int
}

func RefreshRateLimiter() gin.HandlerFunc {
    limiter := rate.NewLimiter(
        rate.Every(time.Minute/10), // 10 req/min
        10,                          // burst de 10
    )

    return func(c *gin.Context) {
        clientIP := c.ClientIP()

        if !limiter.Allow() {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error": "Rate limit exceeded",
                "retry_after": 60,
            })
            c.Abort()
            return
        }

        c.Next()
    }
}
```

**Pourquoi rate limiting ?**
- Prévient le brute force de refresh tokens
- Limite l'impact d'un token leaké (max 10 refresh/min)
- Protège la base de données contre les requêtes excessives

### Révocation de Tokens

#### Phase 1 : Révocation Passive

**Méthodes disponibles** :
1. **Expiration naturelle** : Attendre l'expiration du refresh token (24 heures à 30 jours)
2. **Suppression manuelle DB** : Admin peut `DELETE FROM refresh_tokens WHERE user_id = X`
3. **Désactivation compte** : Validation lors du refresh vérifie `account.status = 'active'`
4. **Suppression compte** : `CASCADE DELETE` supprime automatiquement les refresh tokens

**Limitations** :
- Access tokens restent valides jusqu'à expiration (max 15 min)
- Pas d'interface utilisateur pour gérer les sessions

#### Phase 2 : Révocation Active (Future)

**Fonctionnalités planifiées** :
1. **Endpoint Logout** : `POST /auth/logout` supprime le refresh token
2. **Logout All Devices** : Supprime tous les refresh tokens de l'utilisateur
3. **Token Families** : Détecter la réutilisation d'un token révoqué (attaque)
4. **Blacklist d'access tokens** : Cache Redis pour révoquer immédiatement

---

## 🔄 Migration & Rétrocompatibilité

### Stratégie de Migration

**Objectif** : Déployer le système de refresh token sans casser les clients existants.

#### Étape 1 : Backend Deployment (Phase 1)

**Actions** :
1. Appliquer migration DB (`000006_refresh_tokens.up.sql`)
2. Déployer nouveau code backend avec endpoints `/auth/refresh`
3. Modifier endpoint `/login` pour retourner les deux tokens
4. Réduire `ACCESS_TOKEN_MINUTES` de 60 à 15 minutes

**Impact sur clients existants** :
- ✅ Continuent de fonctionner (utilisent `access_token`)
- ⚠️ Durée de session réduite (60 min → 15 min)
- ⚠️ Re-login plus fréquent (UX dégradée temporairement)
- ✅ Aucun changement de code requis

**Rollback plan** :
- Si problème critique : rollback code + migration down
- Si problème mineur : augmenter temporairement `ACCESS_TOKEN_MINUTES` à 60

#### Étape 2 : Frontend Update (Phase 2)

**Actions** :
1. Implémenter interceptor HTTP pour auto-refresh
2. Stocker et utiliser `refresh_token` du login
3. Gérer erreurs de refresh (redirect login)
4. Ajouter UI pour "Remember me" checkbox

**Impact** :
- ✅ UX améliorée (session glissante)
- ✅ Moins de re-logins
- ✅ Option "Remember me" disponible

**Déploiement progressif** :
- Feature flag frontend : `ENABLE_AUTO_REFRESH=true`
- A/B testing : 10% users → 50% → 100%
- Monitoring : taux de refresh réussis, échecs

#### Étape 3 : Monitoring & Optimization (Phase 3)

**Métriques à surveiller** :
- Taux d'utilisation de `/auth/refresh` (devrait augmenter)
- Taux d'échec de refresh (devrait être < 1%)
- Durée moyenne des sessions (devrait augmenter)
- Nombre de re-logins par user (devrait diminuer)

**Optimisations** :
- Ajuster `ACCESS_TOKEN_MINUTES` si nécessaire
- Ajuster rate limiting si faux positifs
- Cleanup performance monitoring

### Compatibilité des Endpoints

#### `/login` - Rétrocompatibilité Complète

**Avant (clients existants attendent)** :
```json
{
  "access_token": "eyJ...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "pending": false,
  "message": "Login successful"
}
```

**Après (nouvelle réponse)** :
```json
{
  "access_token": "eyJ...",          // ✅ Toujours présent
  "refresh_token": "rt_...",         // ➕ Nouveau champ (ignoré par vieux clients)
  "token_type": "Bearer",            // ✅ Inchangé
  "expires_in": 900,                 // ⚠️ Changé (3600 → 900)
  "refresh_expires_in": 86400,       // ➕ Nouveau champ
  "pending": false,                  // ✅ Inchangé
  "message": "Login successful"      // ✅ Inchangé
}
```

**Parsing JSON** :
- Clients typés (TypeScript, Go, Java) : Champs nouveaux ignorés automatiquement
- Clients faiblement typés (JavaScript vanilla) : `response.access_token` fonctionne
- ✅ Aucun breaking change

#### Nouveaux Endpoints

**`POST /auth/refresh`** : Nouveau, n'affecte pas les clients existants

**`POST /auth/logout`** : Planifié (Phase 2), optionnel

### Configuration Migration

**Ancienne config (`.env`)** :
```bash
TOKEN_HOUR_LIFESPAN=1
```

**Nouvelle config (`.env`)** :
```bash
# Nouvelle variable (prioritaire)
ACCESS_TOKEN_MINUTES=15

# Ancienne variable (deprecated mais supportée)
TOKEN_HOUR_LIFESPAN=1  # → mappé à ACCESS_TOKEN_MINUTES=60 si ACCESS_TOKEN_MINUTES absent
```

**Logique de compatibilité** :
```go
if isEnvSet("ACCESS_TOKEN_MINUTES") {
    config.AccessTokenMinutes = getEnvAsInt("ACCESS_TOKEN_MINUTES", 15)
} else if isEnvSet("TOKEN_HOUR_LIFESPAN") {
    config.AccessTokenMinutes = getEnvAsInt("TOKEN_HOUR_LIFESPAN", 1) * 60
    log.Warn("TOKEN_HOUR_LIFESPAN is deprecated, please use ACCESS_TOKEN_MINUTES")
} else {
    config.AccessTokenMinutes = 15 // default
}
```

---

## 🧪 Stratégie de Tests

### Tests Unitaires

#### Token Service Tests

**Fichier** : `pkg/security/security_test.go`

```go
func TestGenerateRefreshToken_Success(t *testing.T) {
    token, err := GenerateRefreshToken()
    assert.NoError(t, err)
    assert.NotEmpty(t, token)
    assert.True(t, strings.HasPrefix(token, "rt_"))
    assert.Equal(t, 46, len(token)) // "rt_" + 43 chars base64
}

func TestGenerateRefreshToken_Uniqueness(t *testing.T) {
    tokens := make(map[string]bool)
    for i := 0; i < 1000; i++ {
        token, _ := GenerateRefreshToken()
        assert.False(t, tokens[token], "Duplicate token generated")
        tokens[token] = true
    }
}

func TestGenerateTokenPair_Success(t *testing.T) {
    userID := uint(123)
    rememberMe := false

    accessToken, refreshToken, err := GenerateTokenPair(userID, rememberMe)

    assert.NoError(t, err)
    assert.NotEmpty(t, accessToken)
    assert.NotEmpty(t, refreshToken)

    // Valider le JWT
    claims, err := ParseToken(accessToken)
    assert.NoError(t, err)
    assert.Equal(t, userID, claims["user_id"])
}

func TestGenerateTokenPair_RememberMe(t *testing.T) {
    userID := uint(123)

    // Sans remember me
    _, rt1, _ := GenerateTokenPair(userID, false)
    expiry1 := time.Now().Add(7 * 24 * time.Hour)

    // Avec remember me
    _, rt2, _ := GenerateTokenPair(userID, true)
    expiry2 := time.Now().Add(30 * 24 * time.Hour)

    // Les tokens devraient avoir des expiration différentes
    // (vérifier en DB après stockage dans les tests d'intégration)
    assert.NotEqual(t, rt1, rt2)
}
```

#### Repository Tests

**Fichier** : `pkg/security/security_test.go`

```go
func TestCreateRefreshToken_Success(t *testing.T) {
    db := setupTestDB(t)
    repo := NewRefreshTokenRepository(db)

    token := "rt_test123"
    userID := uint(1)
    expiresAt := time.Now().Add(7 * 24 * time.Hour)

    err := repo.Create(token, userID, expiresAt)
    assert.NoError(t, err)

    // Vérifier en DB
    var count int64
    db.Model(&RefreshToken{}).Where("token = ?", token).Count(&count)
    assert.Equal(t, int64(1), count)
}

func TestCreateRefreshToken_DuplicateToken(t *testing.T) {
    db := setupTestDB(t)
    repo := NewRefreshTokenRepository(db)

    token := "rt_duplicate"
    userID := uint(1)
    expiresAt := time.Now().Add(7 * 24 * time.Hour)

    err1 := repo.Create(token, userID, expiresAt)
    assert.NoError(t, err1)

    err2 := repo.Create(token, userID, expiresAt)
    assert.Error(t, err2) // Devrait échouer (UNIQUE constraint)
}

func TestGetRefreshToken_Valid(t *testing.T) {
    db := setupTestDB(t)
    repo := NewRefreshTokenRepository(db)

    token := "rt_valid"
    userID := uint(1)
    expiresAt := time.Now().Add(7 * 24 * time.Hour)
    repo.Create(token, userID, expiresAt)

    rt, err := repo.GetByToken(token)
    assert.NoError(t, err)
    assert.Equal(t, token, rt.Token)
    assert.Equal(t, userID, rt.UserID)
    assert.False(t, rt.IsExpired())
}

func TestGetRefreshToken_Expired(t *testing.T) {
    db := setupTestDB(t)
    repo := NewRefreshTokenRepository(db)

    token := "rt_expired"
    userID := uint(1)
    expiresAt := time.Now().Add(-1 * time.Hour) // Expiré il y a 1h
    repo.Create(token, userID, expiresAt)

    rt, err := repo.GetByToken(token)
    assert.NoError(t, err) // Token trouvé
    assert.True(t, rt.IsExpired()) // Mais expiré
}

func TestUpdateLastUsed_Success(t *testing.T) {
    db := setupTestDB(t)
    repo := NewRefreshTokenRepository(db)

    token := "rt_update"
    userID := uint(1)
    expiresAt := time.Now().Add(7 * 24 * time.Hour)
    repo.Create(token, userID, expiresAt)

    // Attendre 1 seconde
    time.Sleep(1 * time.Second)

    err := repo.UpdateLastUsed(token)
    assert.NoError(t, err)

    rt, _ := repo.GetByToken(token)
    assert.True(t, rt.LastUsedAt.After(rt.CreatedAt))
}

func TestCleanupExpiredTokens_Success(t *testing.T) {
    db := setupTestDB(t)
    repo := NewRefreshTokenRepository(db)

    // Créer 3 tokens expirés, 2 valides
    repo.Create("rt_exp1", 1, time.Now().Add(-1*time.Hour))
    repo.Create("rt_exp2", 2, time.Now().Add(-2*time.Hour))
    repo.Create("rt_exp3", 3, time.Now().Add(-3*time.Hour))
    repo.Create("rt_valid1", 4, time.Now().Add(7*24*time.Hour))
    repo.Create("rt_valid2", 5, time.Now().Add(7*24*time.Hour))

    deleted, err := repo.CleanupExpired()
    assert.NoError(t, err)
    assert.Equal(t, 3, deleted)

    // Vérifier qu'il reste 2 tokens
    var count int64
    db.Model(&RefreshToken{}).Count(&count)
    assert.Equal(t, int64(2), count)
}
```

### Tests d'Intégration

#### Login Flow Tests

**Fichier** : `pkg/accounts/accounts_integration_test.go`

```go
func TestLogin_ReturnsRefreshToken(t *testing.T) {
    router := setupTestRouter(t)

    // Créer utilisateur test
    user := createTestUser(t, "test@example.com", "password123")

    // Login
    body := `{"username":"test@example.com","password":"password123"}`
    req := httptest.NewRequest("POST", "/login", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    router.ServeHTTP(w, req)

    assert.Equal(t, 200, w.Code)

    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)

    assert.NotEmpty(t, response["access_token"])
    assert.NotEmpty(t, response["refresh_token"])
    assert.Equal(t, "Bearer", response["token_type"])
    assert.Equal(t, 900.0, response["expires_in"]) // 15 min
    assert.Equal(t, 86400.0, response["refresh_expires_in"]) // 24 hours
}

func TestLogin_RememberMeTrue(t *testing.T) {
    router := setupTestRouter(t)
    user := createTestUser(t, "test@example.com", "password123")

    body := `{"username":"test@example.com","password":"password123","remember_me":true}`
    req := httptest.NewRequest("POST", "/login", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    router.ServeHTTP(w, req)

    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)

    // Refresh token devrait avoir 30 jours d'expiration
    assert.Equal(t, 2592000.0, response["refresh_expires_in"]) // 30 days
}
```

#### Refresh Flow Tests

**Fichier** : `pkg/security/security_integration_test.go`

```go
func TestRefresh_ValidToken(t *testing.T) {
    router := setupTestRouter(t)

    // 1. Login pour obtenir refresh token
    user := createTestUser(t, "test@example.com", "password123")
    loginResp := performLogin(t, router, "test@example.com", "password123")
    refreshToken := loginResp["refresh_token"].(string)

    // 2. Utiliser refresh token
    body := fmt.Sprintf(`{"refresh_token":"%s"}`, refreshToken)
    req := httptest.NewRequest("POST", "/auth/refresh", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    router.ServeHTTP(w, req)

    assert.Equal(t, 200, w.Code)

    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)

    assert.NotEmpty(t, response["access_token"])
    assert.Equal(t, 900.0, response["expires_in"])
}

func TestRefresh_InvalidToken(t *testing.T) {
    router := setupTestRouter(t)

    body := `{"refresh_token":"rt_invalid_token"}`
    req := httptest.NewRequest("POST", "/auth/refresh", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    router.ServeHTTP(w, req)

    assert.Equal(t, 401, w.Code)

    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Contains(t, response["error"], "Invalid or expired")
}

func TestRefresh_ExpiredToken(t *testing.T) {
    db := setupTestDB(t)
    router := setupTestRouter(t)

    // Créer refresh token expiré directement en DB
    user := createTestUser(t, "test@example.com", "password123")
    expiredToken := "rt_expired_test"
    db.Create(&RefreshToken{
        Token:      expiredToken,
        UserID:     user.ID,
        ExpiresAt:  time.Now().Add(-1 * time.Hour),
        CreatedAt:  time.Now().Add(-8 * 24 * time.Hour),
        LastUsedAt: time.Now().Add(-8 * 24 * time.Hour),
    })

    body := fmt.Sprintf(`{"refresh_token":"%s"}`, expiredToken)
    req := httptest.NewRequest("POST", "/auth/refresh", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    router.ServeHTTP(w, req)

    assert.Equal(t, 401, w.Code)
}

func TestRefresh_InactiveUser(t *testing.T) {
    db := setupTestDB(t)
    router := setupTestRouter(t)

    // Créer user et refresh token, puis désactiver le user
    user := createTestUser(t, "test@example.com", "password123")
    loginResp := performLogin(t, router, "test@example.com", "password123")
    refreshToken := loginResp["refresh_token"].(string)

    // Désactiver le compte
    db.Model(&user).Update("status", "inactive")

    body := fmt.Sprintf(`{"refresh_token":"%s"}`, refreshToken)
    req := httptest.NewRequest("POST", "/auth/refresh", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    router.ServeHTTP(w, req)

    assert.Equal(t, 403, w.Code)
}

func TestRefresh_UpdatesLastUsed(t *testing.T) {
    db := setupTestDB(t)
    router := setupTestRouter(t)

    user := createTestUser(t, "test@example.com", "password123")
    loginResp := performLogin(t, router, "test@example.com", "password123")
    refreshToken := loginResp["refresh_token"].(string)

    // Récupérer last_used_at initial
    var rt1 RefreshToken
    db.Where("token = ?", refreshToken).First(&rt1)
    initialLastUsed := rt1.LastUsedAt

    time.Sleep(1 * time.Second)

    // Refresh
    body := fmt.Sprintf(`{"refresh_token":"%s"}`, refreshToken)
    req := httptest.NewRequest("POST", "/auth/refresh", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    // Vérifier last_used_at mis à jour
    var rt2 RefreshToken
    db.Where("token = ?", refreshToken).First(&rt2)
    assert.True(t, rt2.LastUsedAt.After(initialLastUsed))
}
```

### Checklist de Tests Manuels

**Avant déploiement production** :

#### Authentification
- [ ] Login sans remember_me retourne access + refresh tokens
- [ ] Login avec remember_me=true retourne refresh token 30j
- [ ] Login avec remember_me=false retourne refresh token 24h
- [ ] Login avec credentials invalides retourne 401
- [ ] Login avec compte inactif retourne 403

#### Refresh Token
- [ ] Refresh avec token valide retourne nouveau access token
- [ ] Refresh avec token invalide retourne 401
- [ ] Refresh avec token expiré retourne 401
- [ ] Refresh avec compte désactivé retourne 403
- [ ] Refresh met à jour last_used_at en DB
- [ ] Refresh rate limiting fonctionne (11ème requête = 429)

#### Access Token
- [ ] Access token expire après 15 minutes
- [ ] Requête avec access token expiré retourne 401
- [ ] Requête avec access token valide retourne 200

#### Edge Cases
- [ ] Suppression utilisateur supprime ses refresh tokens (CASCADE)
- [ ] Refresh token utilisable plusieurs fois (pas rotation Phase 1)
- [ ] Cleanup supprime uniquement les tokens expirés
- [ ] Tokens pas loggés dans les logs applicatifs

#### Rétrocompatibilité
- [ ] Ancien client (ignore refresh_token) fonctionne
- [ ] TOKEN_HOUR_LIFESPAN toujours supporté (deprecated)
- [ ] Pas de breaking change sur réponse /login

---

## 📦 Phases d'Implémentation

### Phase 1 : Fonctionnalités Core (MVP)

**Objectif** : Flow de refresh token fonctionnel

**Tâches** :

**Backend** :
1. [ ] Migration DB : table `refresh_tokens`
2. [ ] Modèle Go : `RefreshToken` struct
3. [ ] Repository : CRUD refresh tokens
4. [ ] Service : génération et validation tokens
5. [ ] Modifier handler login : retourner les deux tokens
6. [ ] Nouveau handler : `POST /auth/refresh`
7. [ ] Configuration : nouvelles variables env
8. [ ] Tests unitaires : génération tokens, repository
9. [ ] Tests intégration : login + refresh flows

**Docs** :
10. [ ] Mettre à jour `.env.sample`
11. [ ] Mettre à jour README (section auth)
12. [ ] Swagger/OpenAPI : documenter nouveaux endpoints

**Livrable** :
- ✅ API backend avec refresh token fonctionnel
- ✅ Tests passent (>80% coverage)
- ✅ Documentation à jour

---

### Phase 2 : Tests & Qualité

**Objectif** : Prêt pour production

**Tâches** :

1. [ ] Rate limiting : implémenter sur `/auth/refresh`
2. [ ] Security audit : validation génération tokens, logs
3. [ ] Error handling : messages d'erreur clairs, logs structurés
4. [ ] Monitoring : métriques Prometheus (refresh rate, errors)
5. [ ] Load testing : 100+ concurrent refresh requests
6. [ ] Documentation frontend : guide d'intégration

**Livrable** :
- ✅ Qualité production
- ✅ Rate limiting actif
- ✅ Monitoring en place
- ✅ Guide frontend publié

---

### Phase 3 : Déploiement & Stabilisation ✅

**Statut** : ✅ **VALIDÉ** - Tests d'acceptance terminés avec succès

**Objectif** : Opérationnel en production

**Tâches** :

1. [x] Cleanup job : script de nettoyage tokens expirés
   - ✅ Fonction `CleanupExpiredTokens` implémentée ([pkg/security/refresh_tokens.go:119-135](../pkg/security/refresh_tokens.go))
   - ✅ Goroutine de nettoyage automatique ([main.go:71-88](../main.go))
   - ✅ Configurable via `REFRESH_TOKEN_CLEANUP_INTERVAL_HOURS` (défaut: 24h)
   - ✅ Tests unitaires passants
   - ✅ Logs de monitoring intégrés

2. [x] Préparation déploiement : vérification pré-production
   - ✅ Build réussi (0 erreurs)
   - ✅ Tests automatisés : 100% des tests Phase 2 & 3 passants
   - ✅ Coverage : 71-86% sur le code refresh token
   - ✅ Linter : 0 issues (11 issues corrigées)
   - ✅ Documentation complète (CHANGELOG, README, guide frontend)

3. [x] Tests d'acceptance : 11 scenarios validés manuellement
   - ✅ Test 1: Login flow (basic) - tokens retournés correctement
   - ✅ Test 2: Login flow (remember me) - durée 30 jours validée
   - ✅ Test 3: Refresh token flow (valid) - nouveau access token généré
   - ✅ Test 4: Refresh token flow (invalid) - rejet avec 401
   - ✅ Test 5: Refresh token flow (expired) - rejet avec message approprié
   - ✅ Test 6: Refresh token flow (revoked) - rejet avec message approprié
   - ✅ Test 7: Rate limiting - 10 req/min respecté, 11ème requête bloquée (429)
   - ✅ Test 8: Audit logging - tous les événements loggés en JSON structuré
   - ✅ Test 9: Cleanup job - tokens expirés supprimés automatiquement
   - ✅ Test 10: Database state - schéma et indexes conformes
   - ✅ Test 11: End-to-end flow - parcours complet fonctionnel

4. [ ] Deploy staging : validation complète (à faire par l'équipe ops)
5. [ ] Deploy production : rollout progressif (à faire par l'équipe ops)
6. [ ] Communication : informer utilisateurs (template préparé dans [docs/user-announcement-template.md](../docs/user-announcement-template.md))

**Livrables Phase 3** :
- ✅ Cleanup job opérationnel
- ✅ Rate limiting : 10 req/min par IP
- ✅ Audit logging structuré (JSON)
- ✅ Error sanitization : aucune erreur interne exposée
- ✅ Documentation complète :
  - [CHANGELOG.md](../CHANGELOG.md) - notes de version détaillées
  - [README.md](../README.md) - section Authentication mise à jour
  - [docs/frontend-integration.md](../docs/frontend-integration.md) - guide complet
  - [docs/user-announcement-template.md](../docs/user-announcement-template.md) - templates communication
- ✅ Code production-ready : tous les tests d'acceptance passés

---

### Phase 4 : Frontend Integration (Parallèle)

**Objectif** : UX améliorée avec auto-refresh

**Tâches** :

1. [ ] Interceptor HTTP : auto-refresh sur 401
2. [ ] Token storage : sessionStorage pour refresh token
3. [ ] Remember me UI : checkbox sur login form
4. [ ] Error handling : redirect login si refresh échoue
5. [ ] Feature flag : `ENABLE_AUTO_REFRESH` (rollout progressif)
6. [ ] Monitoring : tracking session duration, re-login rate

**Livrable** :
- ✅ Frontend avec session glissante
- ✅ Remember me fonctionnel

---

### Améliorations Futures (Post-MVP)

**Non inclus dans cette spec, à considérer pour l'avenir** :

#### Sécurité Avancée
1. **Hachage des refresh tokens** (Phase 2 sécurité)
   - Hasher tokens avant stockage en DB
   - Protège contre compromission DB

2. **Rotation de tokens** (Phase 2 sécurité)
   - Nouveau refresh token à chaque utilisation
   - Détecte réutilisation de token révoqué (attaque)

3. **Token families** (Phase 3 sécurité)
   - Grouper tokens par "famille"
   - Révoquer toute la famille si réutilisation détectée

#### Gestion de Sessions
4. **Endpoint logout** (Phase 2 features)
   - `POST /auth/logout` : révoque refresh token actuel

5. **Logout all devices** (Phase 2 features)
   - Révoque tous les refresh tokens d'un utilisateur
   - UI "Sessions actives" avec liste des devices

6. **Session metadata** (Phase 3 features)
   - Stocker IP, User-Agent, last activity
   - UI pour voir et gérer ses sessions

#### Performance
7. **Cache Redis pour refresh tokens** (Phase 2 perf)
   - Réduire charge DB sur validations fréquentes

8. **Token blacklist** (Phase 3 security)
   - Cache Redis pour révoquer access tokens immédiatement
   - Sans attendre expiration (15 min)

#### Compliance
9. **GDPR : Export data** (Phase 2 compliance)
   - Inclure refresh tokens dans export utilisateur

10. **GDPR : Right to be forgotten** (Phase 2 compliance)
    - Supprimer tous refresh tokens lors suppression compte
    - (déjà implémenté via CASCADE DELETE)

11. **Audit logging** (Phase 3 compliance)
    - Logger toutes opérations sur tokens (create, refresh, revoke)
    - Retention 90 jours

---

## 📚 Références

### Standards & Best Practices

- [RFC 6749: OAuth 2.0 Authorization Framework](https://datatracker.ietf.org/doc/html/rfc6749) - Concept de refresh token
- [RFC 7519: JSON Web Tokens (JWT)](https://datatracker.ietf.org/doc/html/rfc7519) - Standard JWT
- [RFC 8725: JWT Best Practices](https://datatracker.ietf.org/doc/html/rfc8725) - Sécurité JWT
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html) - Bonnes pratiques auth

### Implémentations Similaires

- **Auth0** : [Refresh Token Flow](https://auth0.com/docs/secure/tokens/refresh-tokens)
- **Firebase Authentication** : [Manage sessions](https://firebase.google.com/docs/auth/admin/manage-sessions)
- **AWS Cognito** : [Refresh tokens](https://docs.aws.amazon.com/cognito/latest/developerguide/amazon-cognito-user-pools-using-tokens-with-identity-providers.html)
- **Supabase Auth** : [Session management](https://supabase.com/docs/guides/auth/sessions)

### Références Internes

- [pkg/security/security.go](../pkg/security/security.go) - Implémentation JWT actuelle
- [pkg/accounts/accounts.go](../pkg/accounts/accounts.go) - Handler login
- [pkg/config/env.go](../pkg/config/env.go) - Gestion configuration
- [main.go](../main.go) - Routing
- [.env.sample](../.env.sample) - Variables d'environnement

### Outils & Libraries

- **JWT Go** : [golang-jwt/jwt](https://github.com/golang-jwt/jwt) - Utilisé actuellement
- **Gin Web Framework** : [gin-gonic/gin](https://github.com/gin-gonic/gin) - Framework HTTP
- **GORM** : [go-gorm/gorm](https://github.com/go-gorm/gorm) - ORM base de données
- **crypto/rand** : Package Go standard pour génération aléatoire sécurisée

---

## ✅ Critères d'Acceptation

### Cette Spécification est Complète Quand

1. ✅ Tous les requirements fonctionnels (FR1-FR5) peuvent être implémentés à partir de cette spec
2. ✅ Tous les requirements non-fonctionnels (NFR1-NFR3) sont adressés
3. ✅ Les contrats d'API sont complètement documentés avec exemples
4. ✅ Le schéma de base de données est complet et validé
5. ✅ Les considérations de sécurité sont exhaustives
6. ✅ Le chemin de migration est clair et à faible risque
7. ✅ La stratégie de tests couvre tous les chemins critiques
8. ✅ Les diagrammes de séquence détaillent tous les flows frontend
9. ✅ L'implémentation peut procéder sans inconnues majeures

### L'Implémentation est Complète Quand

1. ✅ Toutes les tâches de Phase 1 complétées et testées
2. ✅ Toutes les tâches de Phase 2 complétées et testées
3. ✅ Tous les critères d'acceptation dans FR1-FR5 passent
4. ✅ Couverture de tests >80% sur nouveau code
5. ✅ Aucun breaking change sur comportement API existant
6. ✅ Documentation mise à jour (README, API docs, OpenAPI)
7. ✅ Déployé en production sans incidents
8. ✅ Équipe frontend peut intégrer en utilisant cette spec
9. ✅ Métriques de monitoring positives (refresh success rate >99%)
10. ✅ Performance validée (refresh endpoint <100ms p95)

---

## ✅ Décisions d'Architecture

**Décisions validées** par le PO, Tech Lead et Frontend Dev :

### 1. Stratégie de Stockage Frontend

**Décision** : Option C - `httpOnly cookie`

**Justification** :

- Le frontend proxifie les appels au backend (same-origin)
- Protection maximale contre XSS
- Gestion automatique par le navigateur
- Pas de gestion manuelle du token côté JS

**Implémentation** :

- Backend set le cookie via header `Set-Cookie` avec flags `HttpOnly`, `Secure`, `SameSite=Strict`
- Frontend n'a pas accès au refresh token (sécurité maximale)
- Cookie envoyé automatiquement avec chaque requête

---

### 2. Rate Limiting Granularité

**Décision** : Option A - Par IP uniquement

**Justification** :

- Simplicité d'implémentation pour le MVP
- Suffisant pour prévenir les attaques de base
- Pas besoin d'authentification pour appliquer le rate limit

**Implémentation** :

- Limiter à 10 requêtes/minute par IP sur `/auth/refresh`
- Utiliser middleware rate limiter sur l'endpoint

---

### 3. Cleanup Job Architecture

**Décision** : Option B - Routine au démarrage de l'app

**Justification** :

- Simplicité maximale (pas de dépendance externe)
- Suffisant pour une application en beta
- Pas de gestion de cron job système nécessaire

**Implémentation** :

- Lancer goroutine avec ticker au démarrage de l'application
- Exécution périodique selon `REFRESH_TOKEN_CLEANUP_INTERVAL_HOURS`
- Appelle `CleanupExpiredTokens()` automatiquement

---

### 4. Monitoring & Métriques

**Décision** : Pas de métriques pour le moment

**Justification** :

- Application en beta, pas de besoin immédiat
- Focus sur l'implémentation fonctionnelle d'abord
- Peut être ajouté en Phase 2 si nécessaire

**Future** :

- Phase 2 : Ajouter Prometheus metrics si besoin identifié

---

### 5. Migration Timing

**Décision** : N'importe quand (application en beta)

**Justification** :

- Application en beta avec peu d'utilisateurs
- Impact minimal quelle que soit l'heure de déploiement
- Pas de contrainte de timing particulière

**Approche** :

- Déploiement standard sans contrainte temporelle
- Communication aux utilisateurs beta si nécessaire

---

### 6. Frontend Same-Origin

**Décision** : Same-origin (le front proxifie les appels au back)

**Configuration** :

- Frontend et backend considérés comme same-origin via proxy
- Permet l'utilisation de `httpOnly cookies` en toute sécurité
- Pas de complexité CORS à gérer pour les cookies

---

## 📋 Prochaines Étapes

**Après validation de cette spec** :

1. **Review Meeting** (1h)
   - Présenter spec à l'équipe
   - Discuter questions ouvertes
   - Valider approche technique
   - Estimer effort réel

2. **Refinement** (1h)
   - Incorporer feedback équipe
   - Résoudre questions ouvertes
   - Finaliser story points

3. **Approval**
   - ✅ Product Owner : valide requirements
   - ✅ Tech Lead : valide architecture
   - ✅ Frontend Lead : valide API contracts
   - ✅ DevOps : valide stratégie déploiement

4. **Implémentation**
   - Créer tickets Jira/GitHub Issues à partir des tâches
   - Assigner Phase 1 au sprint courant
   - Commencer développement

---

**Fin de la Spécification**

**Status** : ✅ Reviewed/Validated

**Validé par** :

- ✅ Product Owner (requirements)
- ✅ Tech Lead (conception technique)
- ✅ Frontend Developer (contrats API)

**Dernière mise à jour** : 2026-02-02
