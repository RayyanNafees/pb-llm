# POCKETBASE DOCS|2025-10-15|41 sections

## 1.Introduction
Introduction  Please keep in mind that PocketBase is still under active development and full backward
compatibility is not guaranteed before reaching v1.0.0. PocketBase is NOT recommended for
production critical applications yet, unless you are fine with reading the
and applying some manual migration steps from time to time.
PocketBase is an open source backend consisting of embedded database (SQLite) with realtime subscriptions,
builtin auth management, convenient dashboard UI and simple REST-ish API. It can be used both as Go
framework and as standalone application.
See the
for other platforms and more details.
Once you&#39;ve extracted the archive, you could start the application by running
./pocketbase serve in the extracted directory.
And that&#39;s it!
The first time it will generate an installer link that should be automatically opened in the browser to set
up your first superuser account
(you can also create the first superuser manually via
./pocketbase superuser create EMAIL PASS)
The started web server has the following default routes:
-http://127.0.0.1:8090
- if pb_public directory exists, serves the static content from it (html, css, images,
etc.)
-http://127.0.0.1:8090/_/
- superusers dashboard
-http://127.0.0.1:8090/api/
- REST-ish API
The prebuilt PocketBase executable will create and manage 2 new directories alongside the executable:
-pb_data - stores your application data, uploaded files, etc. (usually should be added in
.gitignore).
-pb_migrations - contains JS migration files with your collection changes (can be safely
committed in your repository).
You can even write custom migration scripts. For more info check the
JS migrations docs.
You could find all available commands and their options by running
./pocketbase --help or
./pocketbase [command] --help

## 2.How to use PocketBase
How to use PocketBase The easiest way to use PocketBase is by interacting with its Web APIs directly from the client-side (e.g.
mobile app or browser SPA).
It was designed with this exact use case in mind and it is also the reason why there are general purpose
JSON APIs for listing, pagination, sorting, filtering, etc.
The access and filter controls for your data are usually done through the
collection API rules
For the cases when you need more specialized handling (sending emails, intercepting the default actions,
creating new routes, etc.) you can
extend PocketBase with Go or JavaScript
For interacting with the
Web APIs
you can make use of the official SDK clients:
-JavaScript SDK (Browser, Node.js, React Native)
-Dart SDK (Web, Mobile, Desktop, CLI)
When used on the client-side, it is safe to have a single/global SDK instance for the entire lifecycle of
your application.
Web apps recommendation    Not everyone will agree with this, but if you are building a web app with PocketBase I recommend
developing the frontend as a traditional client-side SPA and for the cases where additional
server-side handling is needed (e.g. for payment webhooks, extra data server validations, etc.) you
could try to:
-Use PocketBase as Go/JS framework to create new routes or
intercept existing.
-Create one-off Node.js/Bun/Deno/etc. server-side actions that will interact with
PocketBase only as superuser and as pure data store (similar to traditional database
interactions but over HTTP). In this case it is safe to have a global superuser client
such as:
// src/superuser.js
import PocketBase from "pocketbase"
// disable autocancellation so that we can handle async requests from multiple users
superuserClient.autoCancellation(false);
// option 1: authenticate as superuser using email/password (could be filled with ENV params)
await superuserClient.collection('_superusers').authWithPassword(SUPERUSER_EMAIL, SUPERUSER_PASS, {
// This will trigger auto refresh or auto reauthentication in case
autoRefreshThreshold: 30 * 60
// option 2: OR authenticate as superuser via long-lived "API key"
// (see https://pocketbase.io/docs/authentication/#api-keys)
superuserClient.authStore.save('YOUR_GENERATED_SUPERUSER_TOKEN')
export default superuserClient;  Then you can directly import the file in your server-side actions and use the client as
usual:
import superuserClient from './src/superuser.js'
async function serverAction(req, resp) {
... do some extra data validations or handling ...
// send a create request as superuser
await superuserClient.collection('example').create({ ... })
in a JS SSR mode
is possible but it comes with many complications and you need to carefully evaluate whether the cost
of having another backend (PocketBase) alongside your existing one (the Node.js server) is worth it.
JS SSR - issues and recommendations #5313 but some of the common pitfalls are:
-Security issues caused by incorrectly initialized and shared JS SDK instance in a long-running
server-side context.
-OAuth2 integration difficulties related to the server-side only OAuth2 flow (or its mixed
&quot;all-in-one&quot; client-side handling and sharing a cookie with the server-side).
-Proxying realtime connections and essentially duplicating the same thing PocketBase already
does.
-Performance bottlenecks caused by the default single-threaded Node.js process and the
excessive resources utilization due to the server-side rendering and heavy back-and-forth
requests communication between the different layers (client&lt;->Node.js&lt;->PocketBase).
This doesn&#39;t mean that using PocketBase with JS SSR is always a &quot;bad thing&quot; but based on the
dozens reported issues so far I would recommend it only after careful evaluation and only to more
experienced developers that have in-depth understanding of the used tools and their trade-offs. If
you still want to use PocketBase to handle regular users authentication with a JS SSR meta
framework, then you can find some JS SDK examples in the repo&#39;s
JS SSR integration section
Why not htmx, Hotwire/Turbo, Unpoly, etc.    htmx, Hotwire/Turbo, Unpoly and other similar tools are commonly used for building server rendered
applications but unfortunately they don&#39;t play well with the JSON APIs and fully stateless nature
of PocketBase.
It is possible to use them with PocketBase but at the moment I don&#39;t recommend it because we lack
the necessary helpers and utilities for building SSR-first applications, which means that you
might have to create from scratch a lot of things on your own such as middlewares for handling
cookies (and eventually also taking care of CORS and CSRF) or custom authentication
endpoints and access controls (the collection API rules apply only for the builtin JSON routes).
In the future we could eventually provide official SSR support in terms of guides and middlewares
for this use case but again - PocketBase wasn&#39;t designed with this in mind and you may want to
reevaluate the tech stack of your application and switch to a traditional client-side SPA as
mentioned earlier or use a different backend solution that might fit better with your use case.
Mobile apps auth persistence    When building mobile apps with the JavaScript SDK or Dart SDK you&#39;ll have to specify a custom
persistence store if you want to preserve the authentication between the various app activities
and open/close state.
The SDKs come with a helper async storage implementation that allows you to hook any custom
persistent layer (local file, SharedPreferences, key-value based database, etc.). Here is a
minimal PocketBase SDKs initialization for React Native (JavaScript) and Flutter (Dart):
Dart  // Node.js and React Native doesn't have native EventSource implementation
// so in order to use the realtime subscriptions you'll need to load EventSource polyfill,
// for example: npm install react-native-sse --save
import eventsource from 'react-native-sse';
import AsyncStorage from '@react-native-async-storage/async-storage';
import PocketBase, { AsyncAuthStore } from 'pocketbase';
// load the polyfill
global.EventSource = eventsource;
// initialize the async store
const store = new AsyncAuthStore({
save:    async (serialized) => AsyncStorage.setItem('pb_auth', serialized),
initial: AsyncStorage.getItem('pb_auth'),
// initialize the PocketBase client
// (it is OK to have a single/global instance for the duration of your application)
const pb = new PocketBase('http://127.0.0.1:8090', store);
console.log(pb.authStore.record)  import 'package:pocketbase/pocketbase.dart';
import 'package:shared_preferences/shared_preferences.dart';
// for simplicity we are using a simple SharedPreferences instance
// but you can also replace it with its safer EncryptedSharedPreferences alternative
final prefs = await SharedPreferences.getInstance();
// initialize the async store
final store = AsyncAuthStore(
save:    (String data) async => prefs.setString('pb_auth', data),
initial: prefs.getString('pb_auth'),
// initialize the PocketBase client
// (it is OK to have a single/global instance for the duration of your application)
final pb = PocketBase('http://127.0.0.1:8090', authStore: store);
print(pb.authStore.record);      React Native file upload on Android and iOS    At the time of writing, React Native on Android and iOS seems to have a non-standard
FormData implementation and for uploading files on these platforms it requires the following
special object syntax:
uri: "...",
type: "...",
name: "..."
}  Or in other words, you may have to apply a conditional handling similar to:
const data = new FormData();
// result is the resolved promise of ImagePicker.launchImageLibraryAsync
let imageUri = result.assets[0].uri;
if (Platform.OS === 'web') {
const req = await fetch(imageUri);
const blob = await req.blob();
data.append('avatar', blob); // regular File/Blob value
} else {
// the below object format works only on Android and iOS
// (FormData.set() also doesn't seem to be supported so we use FormData.append())
data.append('avatar', {
uri:  imageUri,
type: 'image/*',
name: imageUri.split('/').pop(),
collections, records, authentication, relations, files handling, etc.

## 3.pb-ext - Enhanced PocketBase Server
# pb-ext
Enhanced PocketBase server with monitoring, logging & API docs.
<img width="3840" height="2160" alt="pb-ext" src="https://github.com/user-attachments/assets/af360704-c3d6-4d1f-9b49-80229d6570d2" />
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/magooney-loon/pb-ext)
## Core Features
- **API Schema**: Auto-generates OpenAPI docs UI for your endpoints
- **Cron Tracking**: Logs and monitors scheduled cron jobs
- **System Monitoring**: Real-time CPU, memory, disk, network, and runtime metrics
- **Structured Logging**: Complete logging with error tracking and request tracing
- **Visitor Analytics**: Track visitor stats, page views, device types, and browsers
- **PocketBase Integration**: Uses PocketBase's auth system and styling
## Access
- Admin panel:
```bash
127.0.0.1:8090/_
```
- pb-ext dashboard:
```bash
127.0.0.1:8090/_/_
```
## Quick Start
> 🆕 New to Golang and/or PocketBase? [Read this beginner tutorial](TUTORIAL.md).
```go
package main
import (
"flag"
"log"
app "github.com/magooney-loon/pb-ext/core"
"github.com/pocketbase/pocketbase/core"
func main() {
devMode := flag.Bool("dev", false, "Run in developer mode")
flag.Parse()
initApp(*devMode)
func initApp(devMode bool) {
var opts []app.Option
if devMode {
opts = append(opts, app.InDeveloperMode())
} else {
opts = append(opts, app.InNormalMode())
srv := app.New(opts...)
app.SetupLogging(srv)
registerCollections(srv.App())
registerRoutes(srv.App())
registerJobs(srv.App())
srv.App().OnServe().BindFunc(func(e *core.ServeEvent) error {
app.SetupRecovery(srv.App(), e)
if err := srv.Start(); err != nil {
srv.App().Logger().Error("Fatal application error",
"error", err,
"uptime", srv.Stats().StartTime,
"total_requests", srv.Stats().TotalRequests.Load(),
"active_connections", srv.Stats().ActiveConnections.Load(),
"last_request_time", srv.Stats().LastRequestTime.Load(),
log.Fatal(err)
// Example models in cmd/server/collections.go
// Example routes in cmd/server/routes.go
// Example handlers in cmd/server/handlers.go
// Example cron jobs in cmd/server/jobs.go
// You can restructure Your project as You wish,
// just keep this main.go in cmd/server/main.go
// Consider using the cmd/scripts commands for
// streamlined fullstack dx with +Svelte5kit+
// Ready for a production build deployment?
// https://github.com/magooney-loon/pb-deployer
```
```bash
go mod tidy
go run cmd/scripts/main.go --run-only
```
See `**/*/README.md` for detailed docs.
Having issues with Your API Docs?
```bash
127.0.0.1:8090/api/docs/debug/ast
```

## 4.pb-ext - Scripts Documentation
## [COMMAND SEQUENCES]
### > Standard Development Mode
```
$ go run cmd/scripts/main.go
```
*Builds frontend + starts development server*
### > Full System Installation
```
$ go run cmd/scripts/main.go --install
```
### > Frontend Compilation Only
```
$ go run cmd/scripts/main.go --build-only
```
*Compiles frontend assets without server daemon*
### > Development Server Only
```
$ go run cmd/scripts/main.go --run-only
```
*Starts server daemon, skips build sequence*
### > Production Deployment Build
```
$ go run cmd/scripts/main.go --production
```
*Creates optimized production binary + assets*
### > Test Suite Execution
```
$ go run cmd/scripts/main.go --test-only
```
*Runs comprehensive test suite with coverage reports*
### > Custom Output Directory
```
$ go run cmd/scripts/main.go --production --dist release
```
*Production build with custom target directory*
### > System Help Terminal
```
$ go run cmd/scripts/main.go --help
```
*Displays all available command flags and options*
## [DEPLOYMENT INTEGRATION]
### Automated VPS Deployment via pb-deployer:
```
$ git clone https://github.com/magooney-loon/pb-deployer
$ cd pb-deployer && go run cmd/scripts/main.go --install
```
### pb-deployer Features:
[✓] Automated server provisioning + security hardening
[✓] Zero-downtime deployment cycles with rollback
[✓] Production systemd service management
[✓] Full PocketBase v0.20+ compatibility
## [SYSTEM REQUIREMENTS]
[REQUIRED]
├── Go 1.19+        (backend compilation)
├── Node.js 16+     (frontend build system)
├── npm 8+          (dependency management)
└── Git             (version control)
[OPTIONAL]
└── pb-deployer     (production deployment automation)
## [BUILD PROCESS]
[DEVELOPMENT MODE]
1. System validation    → Check Go/Node/npm availability
2. Dependency install   → npm install + go mod tidy
3. Frontend build       → npm run build
4. Asset deployment     → Copy to pb_public/
5. Server startup       → go run ./cmd/server --dev serve
[PRODUCTION MODE]
1. Environment prep     → Clean dist/ directory
2. Dependency install   → Full dependency resolution
3. Frontend build       → Optimized production build
4. Server compilation   → go build -ldflags="-s -w"
5. Asset packaging      → Create deployment archive
6. Metadata generation  → Build info + package metadata
[TEST MODE]
1. System validation    → Verify test environment
2. Test execution       → Run all test suites
3. Coverage analysis    → Generate coverage reports
4. Report generation    → HTML/JSON/TXT outputs
## [TROUBLESHOOTING]
[ERROR: Command not found]
→ Ensure Go/Node/npm are installed and in system PATH
[ERROR: Frontend build failed]
→ Check package.json and run 'npm install' manually
→ Verify frontend/ directory exists with valid source
[ERROR: Server compilation failed]
→ Run 'go mod tidy' to resolve dependencies
→ Check cmd/server/main.go exists
[ERROR: Permission denied]
→ Ensure write permissions for pb_public/ and dist/

## 5.pb-ext - Collections Implementation
```go
package main
// Collection example
import (
"github.com/pocketbase/pocketbase/core"
// registerCollections sets up all database collections for the application
func registerCollections(app core.App) {
app.OnServe().BindFunc(func(e *core.ServeEvent) error {
if existingCollection != nil {
return nil
// Find users collection for optional relation (v2 auth)
usersCollection, err := app.FindCollectionByNameOrId("users")
if err != nil {
return err
// Add optional user relation
collection.Fields.Add(&core.RelationField{
Name:          "user",
Required:      false, // Optional - v1 routes won't use this
CollectionId:  usersCollection.Id,
CascadeDelete: true,
// Add title field (required)
collection.Fields.Add(&core.TextField{
Name:     "title",
Required: true,
Max:      200,
// Add description field (optional)
collection.Fields.Add(&core.TextField{
Name:     "description",
Required: false,
Max:      1000,
// Add completed field (boolean, default false)
collection.Fields.Add(&core.BoolField{
Name: "completed",
// Add priority field (select)
collection.Fields.Add(&core.SelectField{
Name:   "priority",
Values: []string{"low", "medium", "high"},
// Add auto-date fields
collection.Fields.Add(&core.AutodateField{
Name:     "created",
OnCreate: true,
collection.Fields.Add(&core.AutodateField{
Name:     "updated",
OnCreate: true,
OnUpdate: true,
// Set collection rules - public access for v1
collection.ViewRule = nil   // Public read
collection.CreateRule = nil // Public create
collection.UpdateRule = nil // Public update
collection.DeleteRule = nil // Public delete
// Add indexes
// Save the collection
if err := app.Save(collection); err != nil {
return err
return nil
```

## 6.pb-ext - Handlers Implementation
```go
package main
// API_SOURCE
import (
"encoding/json"
"net/http"
"strconv"
"time"
"github.com/pocketbase/pocketbase/core"
// Request types
Title       string `json:"title"`
Description string `json:"description,omitempty"`
Priority    string `json:"priority,omitempty"` // low, medium, high
Completed   bool   `json:"completed"`
Title       *string `json:"title,omitempty"`
Description *string `json:"description,omitempty"`
Priority    *string `json:"priority,omitempty"`
Completed   *bool   `json:"completed,omitempty"`
// API_DESC Get current server time in multiple formats
// API_TAGS public,utility,time
func timeHandler(c *core.RequestEvent) error {
now := time.Now()
return c.JSON(http.StatusOK, map[string]any{
"time": map[string]string{
"iso":       now.Format(time.RFC3339),
"unix":      strconv.FormatInt(now.Unix(), 10),
"unix_nano": strconv.FormatInt(now.UnixNano(), 10),
"utc":       now.UTC().Format(time.RFC3339),
"server":  "pb-ext",
"version": "1.0.0",
// Check authentication - required for creation
if c.Auth == nil {
return c.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
return c.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid JSON payload"})
if req.Title == "" {
return c.JSON(http.StatusBadRequest, map[string]any{"error": "Title is required"})
// Validate priority if provided
if req.Priority != "" && req.Priority != "low" && req.Priority != "medium" && req.Priority != "high" {
return c.JSON(http.StatusBadRequest, map[string]any{"error": "Priority must be 'low', 'medium', or 'high'"})
// Default priority to medium if not provided
if req.Priority == "" {
req.Priority = "medium"
"title":       req.Title,
"description": req.Description,
"priority":    req.Priority,
"completed":   req.Completed,
// Only set user field if authenticated user is from users collection
// Superusers/admins don't have records in users collection
if c.Auth.Collection().Name == "users" {
// If authenticated as superuser, leave user field empty or handle differently
if err != nil {
return c.JSON(http.StatusInternalServerError, map[string]any{"error": "Collection not found"})
record := core.NewRecord(collection)
if err := c.App.Save(record); err != nil {
return c.JSON(http.StatusCreated, map[string]any{
"id":          record.Id,
"title":       record.GetString("title"),
"description": record.GetString("description"),
"priority":    record.GetString("priority"),
"completed":   record.GetBool("completed"),
"created_at":  record.GetDateTime("created"),
"user_id":     record.GetString("user"),
"created_by":  c.Auth.Collection().Name, // Show if created by user or admin
if err != nil {
return c.JSON(http.StatusInternalServerError, map[string]any{"error": "Collection not found"})
// Build query with optional filters
filter := ""
filterParams := make(map[string]any)
// Filter by completion status if provided
if completed := c.Request.URL.Query().Get("completed"); completed != "" {
if completed == "true" || completed == "1" {
filter = "completed = true"
} else if completed == "false" || completed == "0" {
filter = "completed = false"
// Filter by priority if provided
if priority := c.Request.URL.Query().Get("priority"); priority != "" {
if filter != "" {
filter += " && "
filter += "priority = {:priority}"
filterParams["priority"] = priority
// For authenticated requests, filter by user (only if user is from users collection)
if c.Auth != nil && c.Auth.Collection().Name == "users" {
if filter != "" {
filter += " && "
filter += "user = {:userId}"
filterParams["userId"] = c.Auth.Id
records, err := c.App.FindRecordsByFilter(collection, filter, "-created", 100, 0, filterParams)
if err != nil {
for i, record := range records {
"id":          record.Id,
"title":       record.GetString("title"),
"description": record.GetString("description"),
"priority":    record.GetString("priority"),
"completed":   record.GetBool("completed"),
"created_at":  record.GetDateTime("created"),
"updated_at":  record.GetDateTime("updated"),
// Include user info if available
if userId := record.GetString("user"); userId != "" {
return c.JSON(http.StatusOK, map[string]any{
"filters": map[string]any{
"completed": c.Request.URL.Query().Get("completed"),
"priority":  c.Request.URL.Query().Get("priority"),
if err != nil {
return c.JSON(http.StatusInternalServerError, map[string]any{"error": "Collection not found"})
if err != nil {
// For authenticated requests, check ownership (only enforce for regular users)
if c.Auth != nil && c.Auth.Collection().Name == "users" {
if userID := record.GetString("user"); userID != "" && userID != c.Auth.Id {
return c.JSON(http.StatusForbidden, map[string]any{"error": "Access denied"})
return c.JSON(http.StatusOK, map[string]any{
"id":          record.Id,
"title":       record.GetString("title"),
"description": record.GetString("description"),
"priority":    record.GetString("priority"),
"completed":   record.GetBool("completed"),
"created_at":  record.GetDateTime("created"),
"updated_at":  record.GetDateTime("updated"),
"user_id":     record.GetString("user"),
// Check authentication - required for updates
if c.Auth == nil {
return c.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
if err != nil {
return c.JSON(http.StatusInternalServerError, map[string]any{"error": "Collection not found"})
if err != nil {
if c.Auth.Collection().Name == "users" {
if userID := record.GetString("user"); userID != c.Auth.Id {
if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
return c.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid JSON payload"})
// Apply updates
updates := make(map[string]any)
if req.Title != nil {
if *req.Title == "" {
return c.JSON(http.StatusBadRequest, map[string]any{"error": "Title cannot be empty"})
record.Set("title", *req.Title)
updates["title"] = *req.Title
if req.Description != nil {
record.Set("description", *req.Description)
updates["description"] = *req.Description
if req.Priority != nil {
if *req.Priority != "low" && *req.Priority != "medium" && *req.Priority != "high" {
return c.JSON(http.StatusBadRequest, map[string]any{"error": "Priority must be 'low', 'medium', or 'high'"})
record.Set("priority", *req.Priority)
updates["priority"] = *req.Priority
if req.Completed != nil {
record.Set("completed", *req.Completed)
updates["completed"] = *req.Completed
if err := c.App.Save(record); err != nil {
return c.JSON(http.StatusOK, map[string]any{
"id":          record.Id,
"title":       record.GetString("title"),
"description": record.GetString("description"),
"priority":    record.GetString("priority"),
"completed":   record.GetBool("completed"),
"updated_at":  record.GetDateTime("updated"),
"updates": updates,
// Check authentication - required for deletion
if c.Auth == nil {
return c.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
if err != nil {
return c.JSON(http.StatusInternalServerError, map[string]any{"error": "Collection not found"})
if err != nil {
if c.Auth.Collection().Name == "users" {
if userID := record.GetString("user"); userID != c.Auth.Id {
if err := c.App.Delete(record); err != nil {
return c.JSON(http.StatusOK, map[string]any{
"title": record.GetString("title"),
"deleted_at": time.Now().Format(time.RFC3339),
```

## 7.pb-ext - Jobs Implementation
```go
package main
import (
"fmt"
"time"
"github.com/magooney-loon/pb-ext/core/server"
"github.com/pocketbase/pocketbase/core"
func registerJobs(app core.App) {
app.OnServe().BindFunc(func(e *core.ServeEvent) error {
// Register example cron jobs
if err := helloJob(app); err != nil {
app.Logger().Error("Failed to register hello job", "error", err)
return err
if err := dailyCleanupJob(app); err != nil {
app.Logger().Error("Failed to register daily cleanup job", "error", err)
return err
if err := weeklyStatsJob(app); err != nil {
app.Logger().Error("Failed to register weekly stats job", "error", err)
return err
app.Logger().Info("All cron jobs registered successfully")
func helloJob(app core.App) error {
jobManager := server.GetJobManager()
if jobManager == nil {
return fmt.Errorf("job manager not initialized")
return jobManager.RegisterJob("helloWorld", "Hello World Job",
"A simple demonstration job that runs every 5 minutes, outputs timestamped hello messages and simulates basic task processing",
"*/5 * * * *", func(jobLogger *server.JobExecutionLogger) {
jobLogger.Start("Hello World Job")
jobLogger.Info("Current time: %s", time.Now().Format("2006-01-02 15:04:05"))
jobLogger.Progress("Processing hello world task...")
// Simulate some work
time.Sleep(100 * time.Millisecond)
jobLogger.Success("Hello from cron job! Task completed successfully.")
jobLogger.Complete(fmt.Sprintf("Job finished at: %s", time.Now().Format("2006-01-02 15:04:05")))
func dailyCleanupJob(app core.App) error {
jobManager := server.GetJobManager()
if jobManager == nil {
return fmt.Errorf("job manager not initialized")
return jobManager.RegisterJob("dailyCleanup", "Daily Cleanup Job",
"0 2 * * *", func(jobLogger *server.JobExecutionLogger) {
jobLogger.Start("Daily Cleanup Job")
jobLogger.Info("Cleanup job started at: %s", time.Now().Format("2006-01-02 15:04:05"))
app.Logger().Info("Running daily cleanup job", "time", time.Now())
if err != nil {
jobLogger.Fail(err)
return
cutoffDate := time.Now().AddDate(0, 0, -30)
filter := "completed = true && created < {:cutoff}"
records, err := app.FindRecordsByFilter(collection, filter, "", 100, 0, map[string]any{
"cutoff": cutoffDate.Format("2006-01-02 15:04:05.000Z"),
if err != nil {
jobLogger.Fail(err)
return
deletedCount := 0
for _, record := range records {
if err := app.Delete(record); err != nil {
} else {
deletedCount++
jobLogger.Statistics(map[string]interface{}{
"total_found": len(records),
"deleted":     deletedCount,
"failed":      len(records) - deletedCount,
jobLogger.Complete(fmt.Sprintf("Deleted %d/%d records", deletedCount, len(records)))
app.Logger().Info("Daily cleanup completed", "deleted_records", deletedCount)
func weeklyStatsJob(app core.App) error {
jobManager := server.GetJobManager()
if jobManager == nil {
return fmt.Errorf("job manager not initialized")
return jobManager.RegisterJob("weeklyStats", "Weekly Statistics Job",
"0 0 * * 0", func(jobLogger *server.JobExecutionLogger) {
jobLogger.Start("Weekly Statistics Job")
jobLogger.Info("Generating weekly report for week ending: %s", time.Now().Format("2006-01-02"))
app.Logger().Info("Generating weekly statistics", "time", time.Now())
if err != nil {
jobLogger.Fail(err)
return
weekAgo := time.Now().AddDate(0, 0, -7)
filter := "created >= {:week_ago}"
records, err := app.FindRecordsByFilter(collection, filter, "", 1000, 0, map[string]any{
"week_ago": weekAgo.Format("2006-01-02 15:04:05.000Z"),
if err != nil {
jobLogger.Fail(err)
return
completed := 0
pending := 0
for _, record := range records {
if record.GetBool("completed") {
completed++
} else {
pending++
completionRate := float64(0)
if len(records) > 0 {
completionRate = float64(completed) / float64(len(records)) * 100
// Log statistics using the structured method
stats := map[string]interface{}{
"Completion rate":     fmt.Sprintf("%.1f%%", completionRate),
jobLogger.Info("WEEKLY STATISTICS REPORT")
jobLogger.Statistics(stats)
jobLogger.Complete("Weekly statistics report generated successfully")
app.Logger().Info("Weekly statistics generated",
"completion_rate", completionRate,
```

## 8.pb-ext - Routes Implementation
```go
package main
// API_SOURCE
import (
"github.com/magooney-loon/pb-ext/core/server/api"
"github.com/pocketbase/pocketbase/apis"
"github.com/pocketbase/pocketbase/core"
func registerRoutes(app core.App) {
// Initialize version manager with configs
versionManager := api.InitializeVersionedSystem(createAPIVersions(), "v1")
app.OnServe().BindFunc(func(e *core.ServeEvent) error {
// Get version-specific routers
v1Router, err := versionManager.GetVersionRouter("v1", e)
if err != nil {
return err
v2Router, err := versionManager.GetVersionRouter("v2", e)
if err != nil {
return err
// Register v1 routes
registerV1Routes(v1Router)
// Register v2 routes
registerV2Routes(v2Router)
// Register version management endpoints
versionManager.RegisterWithServer(app)
// createAPIVersions creates version configurations with reduced duplication
func createAPIVersions() map[string]*api.APIDocsConfig {
baseConfig := &api.APIDocsConfig{
Title:       "pb-ext demo api",
Description: "Hello world",
BaseURL:     "http://127.0.0.1:8090/",
Enabled:     true,
// Create v1 config
v1Config := *baseConfig
v1Config.Version = "1.0.0"
v1Config.Status = "stable"
// Create v2 config
v2Config := *baseConfig
v2Config.Version = "2.0.0"
v2Config.Status = "testing"
return map[string]*api.APIDocsConfig{
"v1": &v1Config,
"v2": &v2Config,
// registerV1Routes registers all v1 API routes
func registerV1Routes(router *api.VersionedAPIRouter) {
// Option 1: Manual route registration (explicit control)
prefix := "/api/v1"
// Option 2: CRUD convenience method (less boilerplate)
// Uncomment to use instead of manual registration above:
// v1 := router.SetPrefix("/api/v1")
// }, apis.RequireAuth()) // Auth applied to Create, Update, Patch, Delete
// registerV2Routes registers all v2 API routes
func registerV2Routes(router *api.VersionedAPIRouter) {
// Using prefixed router for cleaner code
v2 := router.SetPrefix("/api/v2")
// Utility routes (no auth required)
v2.GET("/time", timeHandler)
// Future v2 routes can be added here easily:
// v2.GET("/status", statusHandler)
// v2.GET("/health", healthHandler)
```

## 9.Going to production
# Going to production
Going to production  ### Deployment strategies
##### Minimal setup
One of the best PocketBase features is that it&#39;s completely portable. This means that it doesn&#39;t require
any external dependency and
could be deployed by just uploading the executable on your server.
Here is an example of starting a production HTTPS server (auto managed TLS with Let&#39;s Encrypt) on a clean
Ubuntu 22.04 installation.
-Consider the following app directory structure:
myapp/
pb_migrations/
pb_hooks/
pocketbase
-Upload the binary and anything else required by your application to your remote server, for
example using
rsync:
rsync -avz -e ssh /local/path/to/myapp root@YOUR_SERVER_IP:/root/pb
-Start a SSH session with your server:
ssh root@YOUR_SERVER_IP
-Start the executable (specifying a domain name will issue a Let&#39;s encrypt certificate for it)
[root@dev ~]$ /root/pb/pocketbase serve yourdomain.com  Notice that in the above example we are logged in as root which allows us to
bind to the
privileged 80 and 443 ports.
For non-root users usually you&#39;ll need special privileges to be able to do
that. You have several options depending on your OS - authbind,
setcap,
iptables, sysctl, etc. Here is an example using setcap:
[myuser@dev ~]$ sudo setcap 'cap_net_bind_service=+ep' /root/pb/pocketbase
-(Optional) Systemd service
You can skip step 3 and create a Systemd service
to allow your application to start/restart on its own.
Here is an example service file (usually created in
/lib/systemd/system/pocketbase.service):
[Unit]
Description = pocketbase
[Service]
Type             = simple
User             = root
Group            = root
LimitNOFILE      = 4096
Restart          = always
RestartSec       = 5s
StandardOutput   = append:/root/pb/std.log
StandardError    = append:/root/pb/std.log
WorkingDirectory = /root/pb
ExecStart        = /root/pb/pocketbase serve yourdomain.com
[Install]
WantedBy = multi-user.target  After that we just have to enable it and start the service using systemctl:
[root@dev ~]$ systemctl enable pocketbase.service
[root@dev ~]$ systemctl start pocketbase  You can find a link to the Web UI installer in the /root/pb/std.log, but
alternatively you can also create the first superuser explicitly via the
superuser PocketBase command:
[root@dev ~]$ /root/pb/pocketbase superuser create EMAIL PASS
##### Using reverse proxy
If you plan on hosting multiple applications on a single server or need finer network controls, you can
always put PocketBase behind a reverse proxy such as
NGINX, Apache, Caddy, etc.
Just note that when using a reverse proxy you may need to set up the &quot;User IP proxy headers&quot; in the
PocketBase settings so that the application can extract and log the actual visitor/client IP (the
headers are usually X-Real-IP, X-Forwarded-For).
Here is a minimal NGINX example configuration:
server {
listen 80;
client_max_body_size 10M;
location / {
# check http://nginx.org/en/docs/http/ngx_http_upstream_module.html#keepalive
proxy_set_header Connection '';
proxy_http_version 1.1;
proxy_read_timeout 360s;
proxy_set_header Host $host;
proxy_set_header X-Real-IP $remote_addr;
proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
proxy_set_header X-Forwarded-Proto $scheme;
# enable if you are serving under a subpath location
# rewrite /yourSubpath/(.*) /$1  break;
proxy_pass http://127.0.0.1:8090;
}  Corresponding Caddy configuration is:
request_body {
max_size 10MB
reverse_proxy 127.0.0.1:8090 {
transport http {
read_timeout 360s
}  ##### Using Docker
Some hosts (e.g. fly.io) use Docker
for deployments. PocketBase doesn&#39;t have an official Docker image, but you could use the below minimal
Dockerfile as an example:
FROM alpine:latest
ARG PB_VERSION=0.30.2
RUN apk add --no-cache \
unzip \
ca-certificates
# uncomment to copy the local pb_migrations dir into the image
# COPY ./pb_migrations /pb/pb_migrations
# uncomment to copy the local pb_hooks dir into the image
# COPY ./pb_hooks /pb/pb_hooks
EXPOSE 8080
# start PocketBase
CMD ["/pb/pocketbase", "serve", "--http=0.0.0.0:8080"]  To persist your data you need to mount a volume at /pb/pb_data.
For a full example you could check the
&quot;Host for free on Fly.io&quot;
guide.
### Backup and Restore
To backup/restore your application it is enough to manually copy/replace your pb_data
directory
(for transactional safety make sure that the application is not running).
To make things slightly easier, PocketBase v0.16+ comes with builtin backups and restore APIs that could
be accessed from the Dashboard (
Settings > Backups
Backups can be stored locally (default) or in a S3 compatible storage (it is recommended to use a separate bucket only for the backups). The generated backup represents a full snapshot as ZIP archive of your pb_data directory (including
the locally stored uploaded files but excluding any local backups or files uploaded to S3).
During the backup&#39;s ZIP generation the application will be temporary set in read-only mode.
Depending on the size of your pb_data this could be a very slow operation and it is
advised in case of large pb_data (e.g. 2GB+) to consider a different backup strategy
(see an example
backup.sh script
that combines sqlite3 .backup + rsync)
### Recommendations
By default, PocketBase uses the internal Unix sendmail command for sending emails.
While it&#39;s OK for development, it&#39;s not very useful for production, because your emails most likely will get
marked as spam or even fail to deliver.
To avoid deliverability issues, consider using a local SMTP server or an external mail service like
MailerSend,
Brevo,
SendGrid,
Mailgun,
AWS SES, etc.
Once you&#39;ve decided on a mail service, you could configure the PocketBase SMTP settings from the
Dashboard > Settings > Mail settings :
As an additional layer of security you can enable the MFA and OTP options for the _superusers
collection, which will enforce an additional one-time password (email code) requirement when authenticating
as superuser.
In case of email deliverability issues, you can also generate an OTP manually using the
To minimize the risk of API abuse (e.g. excessive auth or record create requests) it is recommended to set
up a rate limiter.
PocketBase v0.23.0+ comes with a simple builtin rate limiter that should cover most of the cases but you
are also free to use any external one via reverse proxy if you need more advanced options.
You can configure the builtin rate limiter from the
Dashboard > Settings > Application:
Unix uses &quot;file descriptors&quot; also for network connections and most systems have a default limit
of ~ 1024.
If your application has a lot of concurrent realtime connections, it is possible that at some point you would
get an error such as: Too many open files.
One way to mitigate this is to check your current account resource limits by running
ulimit -a and find the parameter you want to change. For example, if you want to increase the
open files limit (-n), you could run
ulimit -n 4096 before starting PocketBase.
If you are running in a memory constrained environment, defining the
GOMEMLIMIT
environment variable could help preventing out-of-memory (OOM) termination of your process. It is a &quot;soft limit&quot;
meaning that the memory usage could still exceed it in some situations, but it instructs the GC to be more
&quot;aggressive&quot; and run more often if needed. For example: GOMEMLIMIT=512MiB.
If after GOMEMLIMIT you are still experiencing OOM errors, you can try to enable swap
partitioning (if not already) or open a
Q&amp;A discussion
with some steps to reproduce the error in case it is something that we can improve in PocketBase.
It is fine to ignore the below if you are not sure whether you need it.
By default, PocketBase stores the applications settings in the database as plain JSON text, including the
SMTP password and S3 storage credentials.
While this is not a security issue on its own (PocketBase applications live entirely on a single server
and it is expected only authorized users to have access to your server and application data), in some
situations it may be a good idea to store the settings encrypted in case someone get their hands on your
database file (e.g. from an external stored backup).
To store your PocketBase settings encrypted:
-Create a new environment variable and set a random 32 characters string as its value.
e.g. add
export PB_ENCRYPTION_KEY=&quot;mqnMUCRxQ9BHLB1QIzpsFxzX6IxWtkEU&quot;
in your shell profile file
-Start the application with --encryptionEnv=YOUR_ENV_VAR flag.
e.g. pocketbase serve --encryptionEnv=PB_ENCRYPTION_KEY

## 10.pb-deployer - PocketBase Production Deployment
<div align="center">
<img src="frontend/static/favicon.svg" alt="Logo" width="200">
<h1 align="center">pb-deployer</h1>
<h3 align="center">Automates the lifecycle of deploying PocketBase apps to production</h3>
<a href="https://github.com/magooney-loon/pb-deployer/stargazers"><img src="https://img.shields.io/github/stars/magooney-loon/pb-deployer?style=for-the-badge&color=blue" alt="Stargazers"></a>
<a href="https://github.com/magooney-loon/pb-deployer/graphs/contributors"><img src="https://img.shields.io/github/contributors/magooney-loon/pb-deployer?style=for-the-badge&color=blue" alt="Contributors"></a>
<a href="https://github.com/magooney-loon/pb-deployer/blob/main/LICENSE"><img src="https://img.shields.io/github/license/magooney-loon/pb-deployer?style=for-the-badge&color=blue" alt="AGPL-3.0"></a>
<br>
<img src="frontend/static/deployer.png" alt="Screenshot" width="100%">
<h5 align="center">**WARNING**HOBBY PROJECT**</h5>
<a target="_blank" href="https://magooney.org/">Web Demo UI</a>
</div>
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/magooney-loon/pb-deployer)
## 🚀 Quick Start
```bash
git clone https://github.com/magooney-loon/pb-deployer
cd pb-deployer
go run cmd/scripts/main.go --install
```
## Core Workflow
1. **Server Registration**: Add remote host connection details
2. **Server Setup**: Automated user creation and directory structure
3. **Security Lockdown**: Firewall, fail2ban, disable root SSH (Optional)
4. **App Deployment**: Upload prod dist, systemd service creation
5. **Version Management**: Rollback support with file storage
## Directory Structure
```
/opt/pocketbase/
├── apps/           # Application deployments (per app directory)
├── backups/        # Deployment backups (timestamped)
├── logs/           # Application logs
└── staging/        # Temporary staging during deployments
```
## Deployment Steps
2. **Checking service status**
3. **Stopping existing service**
4. **Creating backup of current deployment**
5. **Preparing deployment directory**
6. **Installing new version**
7. **Creating/updating systemd service**
8. **Creating superuser (if initial deployment)**
9. **Starting service**
10. **Verifying & finalizing deployment**
<div align="center">
<img src="frontend/static/deployer2.png" alt="Logo" width="100%">
</div>
See `**/*/README.md` for detailed docs.
Make sure you loaded your SSH keys, check with `ssh-add -l`
## Contribution

## 11.Introduction - Collections
Response 200:
{"slug:autogenerate":"abc-"}
Response 200:
{"permissions+": "optionA", "roles-": ["staff", "editor"]}
Response 200:
{"documents+": new File(...), "documents-": ["file1_Ab24ZjL.txt", "file2_Frq24ZjL.txt"]}
Response 200:
{"users+": "USER_ID", "categories-": ["CAT_ID1", "CAT_ID2"]}
Response 200:
{"lon":12.34,"lat":56.78}
- Collections
Collections  ### Overview
Collections represent your application data. Under the hood they are backed by plain
SQLite tables that are generated automatically with the collection
name and fields (columns).
Single entry of a collection is called record (a single row in the SQL table).
You can manage your collections from the Dashboard, with the Web APIs using the
client-side SDKs
(superusers only) or programmatically via the
Go/JavaScript
migrations.
Similarly, you can manage your records from the Dashboard, with the Web APIs using the
client-side SDKs
or programmatically via the
Go/JavaScript
Record operations.
Here is what a collection edit panel looks like in the Dashboard:
Currently there are 3 collection types: Base, View and
Auth.
##### Base collection
Base collection is the default collection type and it could be used to store any application
data (articles, products, posts, etc.).
##### View collection
View collection is a read-only collection type where the data is populated from a plain
SQL SELECT statement, allowing users to perform aggregations or any other custom queries in
general.
For example, the following query will create a read-only collection with 3 posts
fields - id, name and totalComments:
SELECT
posts.id,
posts.name,
count(comments.id) as totalComments
FROM posts
LEFT JOIN comments on comments.postId = posts.id
GROUP BY posts.id   View collections don&#39;t receive realtime events because they don&#39;t have create/update/delete
operations. ##### Auth collection
Auth collection has everything from the Base collection but with some additional
special fields to help you manage your app users and also provide various authentication options.
Each Auth collection has the following special system fields:
email, emailVisibility, verified,
password and tokenKey.
They cannot be renamed or deleted but can be configured using their specific field options. For example you
can make the user email required or optional.
You can have as many Auth collections as you want (users, managers, staffs, members, clients, etc.) each
with their own set of fields, separate login and records managing endpoints.
You can build all sort of different access controls:
-Role (Group)  For example, you could attach a &quot;role&quot; select field to your Auth collection with the
following options: &quot;employee&quot; and &quot;staff&quot;. And then in some of your other collections you could
define the following rule to allow only &quot;staff&quot;:
@request.auth.role = &quot;staff&quot;
-Relation (Ownership)  Let&#39;s say that you have 2 collections - &quot;posts&quot; base collection and &quot;users&quot; auth collection. In
your &quot;posts&quot; collection you can create &quot;author&quot;
relation field pointing to the &quot;users&quot; collection. To allow access to only the
&quot;author&quot; of the record(s), you could use a rule like:
@request.auth.id != &quot;&quot; &amp;&amp; author = @request.auth.id
Nested relation fields look ups, including back-relations, are also supported, for example:
someRelField.anotherRelField.author = @request.auth.id
-Managed  In addition to the default &quot;List&quot;, &quot;View&quot;, &quot;Create&quot;, &quot;Update&quot;, &quot;Delete&quot; API rules, Auth
collections have also a special &quot;Manage&quot; API rule that could be used to allow one user (it could
be even from a different collection) to be able to fully manage the data of another user (e.g.
changing their email, password, etc.).
-Mixed  You can build a mixed approach based on your unique use-case. Multiple rules can be grouped with
parenthesis () and combined with &amp;&amp;
(AND) and || (OR) operators:
@request.auth.id != &quot;&quot; &amp;&amp; (@request.auth.role = &quot;staff&quot; || author = @request.auth.id)
### Fields
All collection fields (with exception of the JSONField) are
non-nullable and use a zero-default for their respective type as fallback value
when missing (empty string for text, 0 for number, etc.).
All field specific modifiers are supported both in the Web APIs and via the record Get/Set
methods.
BoolField    BoolField defines bool type field to store a single false
(default) or true value.
NumberField    NumberField defines number type field for storing numeric/float64 value:
0 (default), 2, -1, 1.5.
The following additional set modifiers are available:
-fieldName+
adds number to the already existing record value.
-fieldName-
subtracts number from the already existing record value.
TextField    TextField defines text type field for storing string values:
&quot;&quot; (default), &quot;example&quot;.
The following additional set modifiers are available:
-fieldName:autogenerate
autogenerate a field value if the AutogeneratePattern field option is set.
For example, submitting:
{"slug:autogenerate":"abc-"} will result in &quot;abc-[random]&quot; slug field value.
EmailField    EmailField defines email type field for storing a single email string address:
URLField    URLField defines url type field for storing a single URL string value:
EditorField    EditorField defines editor type field to store HTML formatted text:
&quot;&quot; (default), &lt;p>example&lt;/p>.
DateField    DateField defines date type field to store a single datetime string value:
&quot;&quot; (default), &quot;2022-01-01 00:00:00.000Z&quot;.
All PocketBase dates at the moment follows the RFC3399 format Y-m-d H:i:s.uZ
(e.g. 2024-11-10 18:45:27.123Z).
Dates are compared as strings, meaning that when using the filters with a date field you&#39;ll
have to specify the full datetime string format. For example to target a single day (e.g.
November 19, 2024) you can use something like:
created >= '2024-11-19 00:00:00.000Z' &amp;&amp; created &lt;= '2024-11-19 23:59:59.999Z'
AutodateField    AutodateField defines an autodate type field and it is similar to the DateField but
its value is auto set on record create/update.
This field is usually used for defining timestamp fields like &quot;created&quot; and &quot;updated&quot;.
SelectField    SelectField defines select type field for storing single or multiple string values
from a predefined list.
It is usually intended for handling enums-like values such as
pending/public/private
statuses, simple client/staff/manager/admin roles, etc.
For single select (the MaxSelect option is &lt;= 1)
the field value is a string:
&quot;&quot;, &quot;optionA&quot;.
For multiple select (the MaxSelect option is >= 2)
the field value is an array:
[], [&quot;optionA&quot;, &quot;optionB&quot;].
The following additional set modifiers are available:
-fieldName+
appends one or more values to the existing one.
-+fieldName
prepends one or more values to the existing one.
-fieldName-
subtracts/removes one or more values from the existing one.
For example: {"permissions+": "optionA", "roles-": ["staff", "editor"]}
FileField    FileField defines file type field for managing record file(s).
PocketBase stores in the database only the file name. The file itself is stored either on the
local disk or in S3, depending on your application storage settings.
For single file (the MaxSelect option is &lt;= 1)
the stored value is a string:
&quot;&quot;, &quot;file1_Ab24ZjL.png&quot;.
For multiple file (the MaxSelect option is >= 2)
the stored value is an array:
[], [&quot;file1_Ab24ZjL.png&quot;, &quot;file2_Frq24ZjL.txt&quot;].
The following additional set modifiers are available:
-fieldName+
appends one or more files to the existing field value.
-+fieldName
prepends one or more files to the existing field value.
-fieldName-
deletes one or more files from the existing field value.
For example:
{"documents+": new File(...), "documents-": ["file1_Ab24ZjL.txt", "file2_Frq24ZjL.txt"]}
You can find more detailed information in the
Files upload and handling guide.
RelationField    RelationField defines relation type field for storing single or multiple collection
record references.
For single relation (the MaxSelect option is &lt;= 1)
the field value is a string:
&quot;&quot;, &quot;RECORD_ID&quot;.
For multiple relation (the MaxSelect option is >= 2)
the field value is an array:
[], [&quot;RECORD_ID1&quot;, &quot;RECORD_ID2&quot;].
The following additional set modifiers are available:
-fieldName+
appends one or more ids to the existing one.
-+fieldName
prepends one or more ids to the existing one.
-fieldName-
subtracts/removes one or more ids from the existing one.
For example: {"users+": "USER_ID", "categories-": ["CAT_ID1", "CAT_ID2"]}
JSONField    JSONField defines json type field for storing any serialized JSON value,
including null (default).
GeoPoint    GeoPoint defines geoPoint type field for storing geographic coordinates
(longitude, latitude) as a serialized json object. For example:
{"lon":12.34,"lat":56.78}.
The default/zero value of a geoPoint is the &quot;Null Island&quot;, aka.
{"lon":0,"lat":0}.
When extending PocketBase with Go/JSVM, the geoPoint field value could be set as
types.GeoPoint instance or a regular map with lon and
lat keys:
JavaScript  // set types.GeoPoint
record.Set("address", types.GeoPoint{Lon:12.34, Lat:45.67})
// set map[string]any
record.Set("address", map[string]any{"lon":12.34, "lat":45.67})
// retrieve the field value as types.GeoPoint struct
address := record.GetGeoPoint("address")  record.set("address", {"lon":12.34, "lat":45.67})
const address = record.get("address")

## 12.Introduction - API rules and filters
- API rules and filters
API rules and filters  ### API rules
API Rules are your collection access controls and data filters.
Each collection has 5 rules, corresponding to the specific API action:
-listRule
-viewRule
-createRule
-updateRule
-deleteRule
Auth collections have an additional options.manageRule used to allow one user (it could be even
from a different collection) to be able to fully manage the data of another user (ex. changing their email,
password, etc.).
Each rule could be set to:
-&quot;locked&quot; - aka. null, which means that the action could be performed
only by an authorized superuser
(this is the default)
-Empty string - anyone will be able to perform the action (superusers, authorized users
and guests)
-Non-empty string - only users (authorized or not) that satisfy the rule filter expression
will be able to perform this action
PocketBase API Rules act also as records filter!
Or in other words, you could for example allow listing only the &quot;active&quot; records of your collection,
by using a simple filter expression such as:
status = &quot;active&quot;
(where &quot;status&quot; is a field defined in your Collection).
Because of the above, the API will return 200 empty items response in case a request doesn&#39;t
satisfy a listRule, 400 for unsatisfied createRule and 404 for
unsatisfied viewRule, updateRule and deleteRule.
All rules will return 403 in case they were &quot;locked&quot; (aka. superuser only) and the request client is
not a superuser.
The API Rules are ignored when the action is performed by an authorized superuser (superusers can access everything)!
### Filters syntax
You can find information about the available fields in your collection API rules tab:
There is autocomplete to help guide you while typing the rule filter expression, but in general you have
access to 3 groups of fields:
-Your Collection schema fields
This includes all nested relation fields too, ex.
someRelField.status != &quot;pending&quot;
-@request.*
Used to access the current request data, such as query parameters, body/form fields, authorized user state,
etc.
@request.context - the context where the rule is used (ex.
@request.context != &quot;oauth2&quot;)
The currently supported context values are
default,
oauth2,
otp,
password,
realtime,
protectedFile.
-@request.method - the HTTP request method (ex.
@request.method = &quot;GET&quot;)
-@request.headers.* - the request headers as string values (ex.
@request.headers.x_token = &quot;test&quot;)
Note: All header keys are normalized to lowercase and &quot;-&quot; is replaced with &quot;_&quot; (for
example &quot;X-Token&quot; is &quot;x_token&quot;).
-@request.query.* - the request query parameters as string values (ex.
@request.query.page = &quot;1&quot;)
-@request.auth.* - the current authenticated model (ex.
@request.auth.id != &quot;&quot;)
-@request.body.* - the submitted body parameters (ex.
@request.body.title != &quot;&quot;)
Note: Uploaded files are not part of the @request.body
because they are evaluated separately (this behavior may change in the future).
-@collection.* This filter could be used to target other collections that are not directly related to the current
one (aka. there is no relation field pointing to it) but both shares a common field value, like
for example a category id:
@collection.news.categoryId ?= categoryId &amp;&amp; @collection.news.author ?= @request.auth.id  In case you want to join the same collection multiple times but based on different criteria, you
can define an alias by appending :alias suffix to the collection name.
// see https://github.com/pocketbase/pocketbase/discussions/3805#discussioncomment-7634791
@request.auth.id != "" &amp;&amp;
@collection.courseRegistrations.user ?= id &amp;&amp;
@collection.courseRegistrations:auth.user ?= @request.auth.id &amp;&amp;
@collection.courseRegistrations.courseGroup ?= @collection.courseRegistrations:auth.courseGroup
The syntax basically follows the format
OPERAND OPERATOR OPERAND, where:
-OPERAND - could be any field literal, string (single or double quoted),
number, null, true, false
-OPERATOR - is one of:
= Equal
-!= NOT equal
-> Greater than
->= Greater than or equal
-&lt; Less than
-&lt;= Less than or equal
-~ Like/Contains (if not specified auto wraps the right string OPERAND in a &quot;%&quot; for wildcard
match)
-!~ NOT Like/Contains (if not specified auto wraps the right string OPERAND in a &quot;%&quot; for
wildcard match)
-?= Any/At least one of Equal
-?!= Any/At least one of NOT equal
-?> Any/At least one of Greater than
-?>= Any/At least one of Greater than or equal
-?&lt; Any/At least one of Less than
-?&lt;= Any/At least one of Less than or equal
-?~ Any/At least one of Like/Contains (if not specified auto wraps the right string OPERAND in a &quot;%&quot; for wildcard
match)
-?!~ Any/At least one of NOT Like/Contains (if not specified auto wraps the right string OPERAND in a &quot;%&quot; for
wildcard match)
To group and combine several expressions you can use parenthesis
(...), &amp;&amp; (AND) and || (OR) tokens.
Single line comments are also supported: // Example comment.
### Special identifiers and modifiers
##### @ macros
The following datetime macros are available and can be used as part of the filter expression:
// all macros are UTC based
@now        - the current datetime as string
@second     - @now second number (0-59)
@minute     - @now minute number (0-59)
@hour       - @now hour number (0-23)
@weekday    - @now weekday number (0-6)
@day        - @now day number
@month      - @now month number
@year       - @now year number
@yesterday  - the yesterday datetime relative to @now as string
@tomorrow   - the tomorrow datetime relative to @now as string
@todayStart - beginning of the current day as datetime string
@todayEnd   - end of the current day as datetime string
@monthStart - beginning of the current month as datetime string
@monthEnd   - end of the current month as datetime string
@yearStart  - beginning of the current year as datetime string
@yearEnd    - end of the current year as datetime string  For example:
`@request.body.publicDate >= @now`  ##### :isset modifier
The :isset field modifier is available only for the @request.* fields and can be
used to check whether the client submitted a specific data with the request. Here is for example a rule that
disallows changing a &quot;role&quot; field:
`@request.body.role:isset = false`  Note that @request.body.*:isset at the moment doesn&#39;t support checking for
new uploaded files because they are evaluated separately and cannot be serialized (this behavior may change in the future).
##### :length modifier
The :length field modifier could be used to check the number of items in an array field
(multiple file, select, relation).
Could be used with both the collection schema fields and the @request.body.* fields. For example:
// check example submitted data: {"someSelectField": ["val1", "val2"]}
@request.body.someSelectField:length > 1
// check existing record field length
someRelationField:length = 2  Note that @request.body.*:length at the moment doesn&#39;t support checking
for new uploaded files because they are evaluated separately and cannot be serialized (this behavior may change in the future).
##### :each modifier
The :each field modifier works only with multiple select, file and
relation
type fields. It could be used to apply a condition on each item from the field array. For example:
// check if all submitted select options contain the "create" text
@request.body.someSelectField:each ~ "create"
// check if all existing someSelectField has "pb_" prefix
someSelectField:each ~ "pb_%"  Note that @request.body.*:each at the moment doesn&#39;t support checking for
new uploaded files because they are evaluated separately and cannot be serialized (this behavior may change in the future).
##### :lower modifier
The :lower field modifier could be used to perform lower-case string comparisons. For example:
// check if the submitted lower-cased body "title" field is equal to "test" ("Test", "tEsT", etc.)
@request.body.title:lower = "test"
// match existing records with lower-cased "title" equal to "test" ("Test", "tEsT", etc.)
title:lower ~ "test"  Under the hood it uses the
SQLite LOWER scalar function
and by default works only for ASCII characters, unless the ICU extension is loaded.
##### geoDistance(lonA, latA, lonB, latB)
The geoDistance(lonA, latA, lonB, latB) function could be used to calculate the Haversine distance
between 2 geographic points in kilometres.
The function is intended to be used primarily with the geoPoint field type, but the accepted
arguments could be any plain number or collection field identifier. If the identifier cannot be resolved
and converted to a numeric value, it resolves to null. Note that the
geoDistance function always results in a single row/record value meaning that &quot;any/at-least-one-of&quot;
type of constraint will be applied even if some of its arguments originate from a multiple relation field.
For example:
// offices that are less than 25km from my location (address is a geoPoint field in the offices collection)
geoDistance(address.lon, address.lat, 23.32, 42.69) &lt; 25  #
-Allow only registered users:
@request.auth.id != ""
-Allow only registered users and return records that are either &quot;active&quot; or &quot;pending&quot;:
@request.auth.id != "" &amp;&amp; (status = "active" || status = "pending")
-Allow only registered users who are listed in an allowed_users multi-relation field value:
@request.auth.id != "" &amp;&amp; allowed_users.id ?= @request.auth.id
-Allow access by anyone and return only the records where the title field value starts with
title ~ "Lorem%"

## 13.Introduction - Working with relations
- Working with relations
Working with relations  ### Overview
Let&#39;s assume that we have the following collections structure:
The relation fields follow the same rules as any other collection field and can be set/modified
by directly updating the field value - with a record id or array of ids, in case a multiple relation is used.
Below is an example that shows creating a new posts record with 2 assigned tags.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
const post = await pb.collection('posts').create({
'tags':  ['TAG_ID1', 'TAG_ID2'],
});  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
final post = await pb.collection('posts').create(body: {
'tags':  ['TAG_ID1', 'TAG_ID2'],
});   ### Prepend/Append to multiple relation
To prepend/append a single or multiple relation id(s) to an existing value you can use the
+ field modifier:
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
const post = await pb.collection('posts').update('POST_ID', {
// prepend single tag
'+tags': 'TAG_ID1',
// append multiple tags at once
'tags+': ['TAG_ID1', 'TAG_ID2'],
})  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
final post = await pb.collection('posts').update('POST_ID', body: {
// prepend single tag
'+tags': 'TAG_ID1',
// append multiple tags at once
'tags+': ['TAG_ID1', 'TAG_ID2'],
})   ### Remove from multiple relation
To remove a single or multiple relation id(s) from an existing value you can use the
- field modifier:
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
const post = await pb.collection('posts').update('POST_ID', {
// remove single tag
'tags-': 'TAG_ID1',
// remove multiple tags at once
'tags-': ['TAG_ID1', 'TAG_ID2'],
})  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
final post = await pb.collection('posts').update('POST_ID', body: {
// remove single tag
'tags-': 'TAG_ID1',
// remove multiple tags at once
'tags-': ['TAG_ID1', 'TAG_ID2'],
})   ### Expanding relations
You can also expand record relation fields directly in the returned response without making additional
requests by using the expand query parameter, e.g. ?expand=user,post.tags
Only the relations that the request client can View (aka. satisfies the relation
collection&#39;s View API Rule) will be expanded.
Nested relation references in expand, filter or sort are supported
via dot-notation and up to 6-levels depth.
For example, to list all comments with their user relation expanded, we can
do the following:
Dart  `await pb.collection("comments").getList(1, 30, { expand: "user" })`  `await pb.collection("comments").getList(perPage: 30, expand: "user")`   {
"page": 1,
"perPage": 30,
"totalPages": 1,
"totalItems": 20,
"items": [
"id": "lmPJt4Z9CkLW36z",
"collectionId": "BHKW36mJl3ZPt6z",
"collectionName": "comments",
"created": "2022-01-01 01:00:00.456Z",
"updated": "2022-01-01 02:15:00.456Z",
"post": "WyAw4bDrvws6gGl",
"user": "FtHAW9feB5rze7D",
"message": "Example message...",
"expand": {
"user": {
"id": "FtHAW9feB5rze7D",
"collectionId": "srmAo0hLxEqYF7F",
"collectionName": "users",
"created": "2022-01-01 00:00:00.000Z",
"updated": "2022-01-01 00:00:00.000Z",
"username": "users54126",
"verified": false,
"emailVisibility": false,
"name": "John Doe"
}  ### Back-relations
PocketBase supports also filter, sort and expand for
back-relations
- relations where the associated relation field is not in the main collection.
The following notation is used: referenceCollection_via_relField (ex.
comments_via_post).
For example, let&#39;s list the posts that have at least one comments record
containing the word &quot;hello&quot;:
Dart  await pb.collection("posts").getList(1, 30, {
filter: "comments_via_post.message ?~ 'hello'"
expand: "comments_via_post.user",
})  await pb.collection("posts").getList(
perPage: 30,
filter: "comments_via_post.message ?~ 'hello'"
expand: "comments_via_post.user",
)   {
"page": 1,
"perPage": 30,
"totalPages": 2,
"totalItems": 45,
"items": [
"id": "WyAw4bDrvws6gGl",
"collectionName": "posts",
"created": "2022-01-01 01:00:00.456Z",
"updated": "2022-01-01 02:15:00.456Z",
"expand": {
"comments_via_post": [
"id": "lmPJt4Z9CkLW36z",
"collectionId": "BHKW36mJl3ZPt6z",
"collectionName": "comments",
"created": "2022-01-01 01:00:00.456Z",
"updated": "2022-01-01 02:15:00.456Z",
"post": "WyAw4bDrvws6gGl",
"user": "FtHAW9feB5rze7D",
"expand": {
"user": {
"id": "FtHAW9feB5rze7D",
"collectionId": "srmAo0hLxEqYF7F",
"collectionName": "users",
"created": "2022-01-01 00:00:00.000Z",
"updated": "2022-01-01 00:00:00.000Z",
"username": "users54126",
"verified": false,
"emailVisibility": false,
"name": "John Doe"
"id": "tu4Z9CkLW36mPJz",
"collectionId": "BHKW36mJl3ZPt6z",
"collectionName": "comments",
"created": "2022-01-01 01:10:00.123Z",
"updated": "2022-01-01 02:39:00.456Z",
"post": "WyAw4bDrvws6gGl",
"user": "FtHAW9feB5rze7D",
"message": "hello...",
"expand": {
"user": {
"id": "FtHAW9feB5rze7D",
"collectionId": "srmAo0hLxEqYF7F",
"collectionName": "users",
"created": "2022-01-01 00:00:00.000Z",
"updated": "2022-01-01 00:00:00.000Z",
"username": "users54126",
"verified": false,
"emailVisibility": false,
"name": "John Doe"
}  ###### Back-relation caveats
-By default the back-relation reference is resolved as a dynamic
multiple relation field, even when the back-relation field itself is marked as
single.
This is because the main record could have more than one single
back-relation reference (see in the above example that the comments_via_post
expand is returned as array, although the original comments.post field is a
single relation).
The only case where the back-relation will be treated as a single
relation field is when there is
UNIQUE index constraint defined on the relation field.
-Back-relation expand is limited to max 1000 records per relation field. If you
need to fetch larger number of back-related records a better approach could be to send a
separate paginated getList() request to the back-related collection to avoid transferring
large JSON payloads and to reduce the memory usage.

## 14.Introduction - Authentication
- Authentication
Authentication  ### Overview
A single client is considered authenticated as long as it sends valid
Authorization:YOUR_AUTH_TOKEN header with the request.
The PocketBase Web APIs are fully stateless and there are no sessions in the traditional sense (even the
tokens are not stored in the database).
Because there are no sessions and we don&#39;t store the tokens on the server there is also no logout
endpoint. To &quot;logout&quot; a user you can simply disregard the token from your local state (aka.
pb.authStore.clear() if you use the SDKs).
The auth token could be generated either through the specific auth collection Web APIs or programmatically
via Go/JS.
All allowed auth collection methods can be configured individually from the specific auth collection
options.
Note that PocketBase admins (aka. _superusers) are similar to the regular auth
collection records with 2 caveats:
-OAuth2 is not supported as auth method for the _superusers collection
-Superusers can access and modify anything (collection API rules are ignored)
### Authenticate with password
To authenticate with password you must enable the Identity/Password auth collection option
(see also
Web API reference
The default identity field is the email but you can configure any other unique field like
&quot;username&quot; (it must have a UNIQUE index).
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
// after the above you can also access the auth data from the authStore
console.log(pb.authStore.isValid);
console.log(pb.authStore.token);
console.log(pb.authStore.record.id);
// "logout" the last authenticated record
pb.authStore.clear();  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
// after the above you can also access the auth data from the authStore
print(pb.authStore.isValid);
print(pb.authStore.token);
print(pb.authStore.record.id);
// "logout" the last authenticated record
pb.authStore.clear();   ### Authenticate with OTP
To authenticate with email code you must enable the One-time password (OTP)
auth collection option
(see also
Web API reference
The usual flow is the user typing manually the received password from their email but you can also
adjust the default email template from the collection options and add a url containing the OTP and its
id as query parameters
Note that when requesting an OTP we return an otpId even if a user with the provided email
doesn&#39;t exist as a very rudimentary enumeration protection (it doesn&#39;t create or send anything).
On successful OTP validation, by default the related user email will be automatically marked as
&quot;verified&quot;.
Keep in mind that OTP as a standalone authentication method could be less secure compared to the
other methods because the generated password is usually 0-9 digits and there is a risk of it being
guessed or enumerated (especially when a longer duration time is configured).
For security critical applications OTP is recommended to be used in combination with the other
auth methods and the Multi-factor authentication option.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
// send OTP email to the provided auth record
// ... show a screen/popup to enter the password from the email ...
// authenticate with the requested OTP id and the email password
const authData = await pb.collection('users').authWithOTP(result.otpId, "YOUR_OTP");
// after the above you can also access the auth data from the authStore
console.log(pb.authStore.isValid);
console.log(pb.authStore.token);
console.log(pb.authStore.record.id);
// "logout"
pb.authStore.clear();  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
// send OTP email to the provided auth record
// ... show a screen/popup to enter the password from the email ...
// authenticate with the requested OTP id and the email password
final authData = await pb.collection('users').authWithOTP(result.otpId, "YOUR_OTP");
// after the above you can also access the auth data from the authStore
print(pb.authStore.isValid);
print(pb.authStore.token);
print(pb.authStore.record.id);
// "logout"
pb.authStore.clear();   ### Authenticate with OAuth2
You can also authenticate your users with an OAuth2 provider (Google, GitHub, Microsoft, etc.). See the
section below for example integrations.
Before starting, you&#39;ll need to create an OAuth2 app in the provider&#39;s dashboard in order to get a
Client Id and Client Secret, and register a redirect URL
Once you have obtained the Client Id and Client Secret, you can
enable and configure the provider from your PocketBase auth collection options (PocketBase > Collections > {YOUR_COLLECTION} > Edit collection (settings cogwheel) > Options
> OAuth2).
This method handles everything within a single call without having to define custom redirects,
deeplinks or even page reload.
When creating your OAuth2 app, for a callback/redirect URL you have to use the
https://yourdomain.com/api/oauth2-redirect
(or when testing locally - http://127.0.0.1:8090/api/oauth2-redirect ).
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('https://pocketbase.io');
// This method initializes a one-off realtime subscription and will
// open a popup window with the OAuth2 vendor page to authenticate.
// Once the external OAuth2 sign-in/sign-up flow is completed, the popup
// window will be automatically closed and the OAuth2 data sent back
// If the popup is being blocked on Safari, make sure that your click handler is not using async/await.
pb.collection('users').authWithOAuth2({
provider: 'google'
}).then((authData) => {
console.log(authData)
// after the above you can also access the auth data from the authStore
console.log(pb.authStore.isValid);
console.log(pb.authStore.token);
console.log(pb.authStore.record.id);
// "logout" the last authenticated record
pb.authStore.clear();
});  import 'package:pocketbase/pocketbase.dart';
import 'package:url_launcher/url_launcher.dart';
final pb = PocketBase('https://pocketbase.io');
// This method initializes a one-off realtime subscription and will
// call the provided urlCallback with the OAuth2 vendor url to authenticate.
// Once the external OAuth2 sign-in/sign-up flow is completed, the browser
// window will be automatically closed and the OAuth2 data sent back
// Note that it requires the app and realtime connection to remain active in the background!
// For Android 15+ check the note in https://github.com/pocketbase/dart-sdk#oauth2-and-android-15.
final authData = await pb.collection('users').authWithOAuth2('google', (url) async {
// or use flutter_custom_tabs to make the transitions between native and web content more seamless
await launchUrl(url);
// after the above you can also access the auth data from the authStore
print(pb.authStore.isValid);
print(pb.authStore.token);
print(pb.authStore.record.id);
// "logout" the last authenticated record
pb.authStore.clear();   When authenticating manually with OAuth2 code you&#39;ll need 2 endpoints:
-somewhere to show the &quot;Login with ...&quot; links
-somewhere to handle the provider&#39;s redirect in order to exchange the auth code for token
Here is a simple web example:
-Links page
(e.g. https://127.0.0.1:8090 serving pb_public/index.html):
&lt;!DOCTYPE html>
&lt;html>
&lt;head>
&lt;meta charset="utf-8" />
&lt;meta name="viewport" content="width=device-width, initial-scale=1" />
&lt;title>OAuth2 links page&lt;/title>
&lt;script src="https://code.jquery.com/jquery-3.7.1.slim.min.js">&lt;/script>
&lt;/head>
&lt;body>
&lt;ul id="list">
&lt;li>Loading OAuth2 providers...&lt;/li>
&lt;/ul>
&lt;script type="module">
import PocketBase from "https://cdn.jsdelivr.net/gh/pocketbase/js-sdk@master/dist/pocketbase.es.mjs"
const pb          = new PocketBase("http://127.0.0.1:8090");
const redirectURL = "http://127.0.0.1:8090/redirect.html";
const authMethods = await pb.collection("users").listAuthMethods();
const providers   = authMethods.oauth2?.providers || [];
const listItems   = [];
for (const provider of providers) {
const $li = $(`&lt;li>&lt;a>Login with ${provider.name}&lt;/a>&lt;/li>`);
$li.find("a")
.attr("href", provider.authURL + redirectURL)
.data("provider", provider)
.click(function () {
// store provider's data on click for verification in the redirect page
localStorage.setItem("provider", JSON.stringify($(this).data("provider")));
listItems.push($li);
$("#list").html(listItems.length ? listItems : "&lt;li>No OAuth2 providers.&lt;/li>");
&lt;/script>
&lt;/body>
&lt;/html>
-Redirect handler page
(e.g. https://127.0.0.1:8090/redirect.html serving
pb_public/redirect.html):
&lt;!DOCTYPE html>
&lt;html>
&lt;head>
&lt;meta charset="utf-8">
&lt;title>OAuth2 redirect page&lt;/title>
&lt;/head>
&lt;body>
&lt;pre id="content">Authenticating...&lt;/pre>
&lt;script type="module">
import PocketBase from "https://cdn.jsdelivr.net/gh/pocketbase/js-sdk@master/dist/pocketbase.es.mjs"
const pb          = new PocketBase("http://127.0.0.1:8090");
const redirectURL = "http://127.0.0.1:8090/redirect.html";
const contentEl   = document.getElementById("content");
// parse the query parameters from the redirected url
const params = (new URL(window.location)).searchParams;
const provider = JSON.parse(localStorage.getItem("provider"))
// compare the redirect's state param and the stored provider's one
if (provider.state !== params.get("state")) {
contentEl.innerText = "State parameters don't match.";
} else {
// authenticate
pb.collection("users").authWithOAuth2Code(
provider.name,
params.get("code"),
provider.codeVerifier,
redirectURL,
// pass any optional user create data
emailVisibility: false,
).then((authData) => {
contentEl.innerText = JSON.stringify(authData, null, 2);
}).catch((err) => {
contentEl.innerText = "Failed to exchange code.\n" + err;
&lt;/script>
&lt;/body>
&lt;/html>
When using the &quot;Manual code exchange&quot; flow for sign-in with Apple your redirect
handler must accept POST requests in order to receive the name and the
email of the Apple user. If you just need the Apple user id, you can keep the redirect
handler GET but you&#39;ll need to replace in the Apple authorization url
response_mode=form_post with response_mode=query.
### Multi-factor authentication
PocketBase v0.23+ introduced optional Multi-factor authentication (MFA).
If enabled, it requires the user to authenticate with any 2 different auth methods from above (the
order doesn&#39;t matter).
The expected flow is:
-User authenticates with &quot;Auth method A&quot;.
-On success, a 401 response is sent with {"mfaId": "..."} as JSON body (the MFA
&quot;session&quot; is stored in the _mfas system collection).
-User authenticates with &quot;Auth method B&quot; as usual
-On success, a regular auth response is returned, aka. token + auth record data.
Below is an example for email/password + OTP authentication:
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
try {
} catch (err) {
const mfaId = err.response?.mfaId;
if (!mfaId) {
throw err; // not mfa -> rethrow
// the user needs to authenticate again with another auth method, for example OTP
// ... show a modal for users to check their email and to enter the received code ...
await pb.collection('users').authWithOTP(result.otpId, 'EMAIL_CODE', { 'mfaId': mfaId });
}  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
try {
} on ClientException catch (e) {
final mfaId = e.response['mfaId'];
if (mfaId == null) {
throw e; // not mfa -> rethrow
// the user needs to authenticate again with another auth method, for example OTP
// ... show a modal for users to check their email and to enter the received code ...
await pb.collection('users').authWithOTP(result.otpId, 'EMAIL_CODE', query: { 'mfaId': mfaId });
}   ### Users impersonation
Superusers have the option to generate tokens and authenticate as anyone else via the
Impersonate endpoint
The generated impersonate auth tokens can have custom duration but are not renewable!
For convenience the official SDKs creates and returns a standalone client that keeps the token state
in memory, aka. only for the duration of the impersonate client instance.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
// authenticate as superuser
// impersonate
// (the custom token duration is in seconds and it is optional)
const impersonateClient = await pb.collection("users").impersonate("USER_RECORD_ID", 3600)
// log the impersonate token and user data
console.log(impersonateClient.authStore.token);
console.log(impersonateClient.authStore.record);
// send requests as the impersonated user
const items = await impersonateClient.collection("example").getFullList();  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
// authenticate as superuser
// impersonate
// (the custom token duration is in seconds and it is optional)
final impersonateClient = await pb.collection("users").impersonate("USER_RECORD_ID", 3600)
// log the impersonate token and user data
print(impersonateClient.authStore.token);
print(impersonateClient.authStore.record);
// send requests as the impersonated user
final items = await impersonateClient.collection("example").getFullList();   ### API keys
While PocketBase doesn&#39;t have &quot;API keys&quot; in the traditional sense, as a side effect of the support for
users impersonation, for such cases you can use instead the generated nonrenewable
_superusers impersonate auth token.
You can generate such token via the above impersonate API or from the
Dashboard &gt; Collections &gt; _superusers &gt; {select superuser} &gt; &quot;Impersonate&quot; dropdown option:
Because of the security implications (superusers can execute, access and modify anything), use the
generated _superusers tokens with extreme care and only for internal
server-to-server communication.
To invalidate already issued tokens, you need to change the individual superuser account password
(or if you want to reset the tokens for all superusers - change the shared auth token secret from
the _superusers collection options).
### Auth token verification
PocketBase doesn&#39;t have a dedicated token verification endpoint, but if you want to verify an existing
auth token from a 3rd party app you can send an
Auth refresh
call, aka. pb.collection(&quot;users&quot;).authRefresh().
On valid token - it returns a new token with refreshed exp claim and the latest user data.
Otherwise - returns an error response.
the new one if you don&#39;t need it (as mentioned in the beginning - PocketBase doesn&#39;t store the tokens on the
server).
Performance wise, the used HS256 algorithm for generating the JWT has very little to no
impact and it is essentially the same in terms of response time as calling
getOne(&quot;USER_ID&quot;) (see
benchmarks

## 15.Introduction - Files upload and handling
- Files upload and handling
Files upload and handling  ### Uploading files
To upload files, you must first add a file field to your collection:
Once added, you could create/update a Record and upload &quot;documents&quot; files by sending a
multipart/form-data request using the Records create/update APIs.
Each uploaded file will be stored with the original filename (sanitized) and suffixed with a
random part (usually 10 characters). For example test_52iwbgds7l.png.
The max allowed size of a single file currently is limited to ~8GB (253-1 bytes).
Here is an example how to create a new record and upload multiple files to the example &quot;documents&quot;
file field using the SDKs:
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
// create a new record and upload multiple files
// (files must be Blob or File instances)
const createdRecord = await pb.collection('example').create({
title: 'Hello world!', // regular text field
'documents': [
new File(['content 1...'], 'file1.txt'),
new File(['content 2...'], 'file2.txt'),
// Alternative FormData + plain HTML file input example
// &lt;input type="file" id="fileInput" />
const fileInput = document.getElementById('fileInput');
const formData = new FormData();
// set regular text field
formData.append('title', 'Hello world!');
// listen to file input changes and add the selected files to the form data
fileInput.addEventListener('change', function () {
for (let file of fileInput.files) {
formData.append('documents', file);
// upload and create new record
const createdRecord = await pb.collection('example').create(formData);  import 'package:pocketbase/pocketbase.dart';
import 'package:http/http.dart' as http;
final pb = PocketBase('http://127.0.0.1:8090');
// create a new record and upload multiple files
final record = await pb.collection('example').create(
body: {
'title': 'Hello world!', // regular text field
files: [
http.MultipartFile.fromString(
'documents',
'example content 1...',
filename: 'file1.txt',
http.MultipartFile.fromString(
'documents',
'example content 2...',
filename: 'file2.txt',
);   If your file field supports uploading multiple files (aka.
Max Files option is &gt;= 2) you can use the + prefix/suffix field name modifier
to respectively prepend/append new files alongside the already uploaded ones. For example:
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
const createdRecord = await pb.collection('example').update('RECORD_ID', {
"documents+": new File(["content 3..."], "file3.txt")
});  import 'package:pocketbase/pocketbase.dart';
import 'package:http/http.dart' as http;
final pb = PocketBase('http://127.0.0.1:8090');
final record = await pb.collection('example').update(
'RECORD_ID',
files: [
http.MultipartFile.fromString(
'documents+',
'example content 3...',
filename: 'file3.txt',
);   ### Deleting files
To delete uploaded file(s), you could either edit the Record from the Dashboard, or use the API and set
the file field to a zero-value  (empty string, []).
If you want to delete individual file(s) from a multiple file upload field, you could
suffix the field name with - and specify the filename(s) you want to delete. Here are some examples
using the SDKs:
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
// delete all "documents" files
await pb.collection('example').update('RECORD_ID', {
'documents': [],
// delete individual files
await pb.collection('example').update('RECORD_ID', {
'documents-': ["file1.pdf", "file2.txt"],
});  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
// delete all "documents" files
await pb.collection('example').update('RECORD_ID', body: {
'documents': [],
// delete individual files
await pb.collection('example').update('RECORD_ID', body: {
'documents-': ["file1.pdf", "file2.txt"],
});   The above examples use the JSON object data format, but you could also use FormData instance
for multipart/form-data requests. If using
FormData set the file field to an empty string.
### File URL
Each uploaded file could be accessed by requesting its file url:
http://127.0.0.1:8090/api/files/COLLECTION_ID_OR_NAME/RECORD_ID/FILENAME
If your file field has the Thumb sizes option, you can get a thumb of the image file by
adding the thumb
query parameter to the url like this:
http://127.0.0.1:8090/api/files/COLLECTION_ID_OR_NAME/RECORD_ID/FILENAME?thumb=100x300  Currently limited to jpg, png, gif (its first frame) and partially webp (stored as png).
The following thumb formats are currently supported:
-WxH
(e.g. 100x300) - crop to WxH viewbox (from center)
-WxHt
(e.g. 100x300t) - crop to WxH viewbox (from top)
-WxHb
(e.g. 100x300b) - crop to WxH viewbox (from bottom)
-WxHf
(e.g. 100x300f) - fit inside a WxH viewbox (without cropping)
-0xH
(e.g. 0x300) - resize to H height preserving the aspect ratio
-Wx0
(e.g. 100x0) - resize to W width preserving the aspect ratio
The original file would be returned, if the requested thumb size is not found or the file is not an image!
If you already have a Record model instance, the SDKs provide a convenient method to generate a file url
by its name.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
const record = await pb.collection('example').getOne('RECORD_ID');
// get only the first filename from "documents"
// note:
// "documents" is an array of filenames because
// the "documents" field was created with "Max Files" option > 1;
// if "Max Files" was 1, then the result property would be just a string
const firstFilename = record.documents[0];
// returns something like:
// http://127.0.0.1:8090/api/files/example/kfzjt5oy8r34hvn/test_52iWbGinWd.png?thumb=100x250
const url = pb.files.getURL(record, firstFilename, {'thumb': '100x250'});  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
final record = await pb.collection('example').getOne('RECORD_ID');
// get only the first filename from "documents"
// note:
// "documents" is an array of filenames because
// the "documents" field was created with "Max Files" option > 1;
// if "Max Files" was 1, then the result property would be just a string
final firstFilename = record.getListValue&lt;String>('documents')[0];
// returns something like:
// http://127.0.0.1:8090/api/files/example/kfzjt5oy8r34hvn/test_52iWbGinWd.png?thumb=100x250
### Protected files
By default all files are publicly accessible if you know their full url.
For most applications this is fine and reasonably safe because all files have a random part appended to
their name, but in some cases you may want an extra security to prevent unauthorized access to sensitive
files like ID card or Passport copies, contracts, etc.
To do this you can mark the file field as Protected from its field options in the
Dashboard and then request the file with a special short-lived file token.
Only requests that satisfy the View API rule of the record collection will be able
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
// authenticate
// generate a file token
const fileToken = await pb.files.getToken();
// retrieve an example protected file url (will be valid ~2min)
const record = await pb.collection('example').getOne('RECORD_ID');
const url = pb.files.getURL(record, record.myPrivateFile, {'token': fileToken});  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
// authenticate
// generate a file token
final fileToken = await pb.files.getToken();
// retrieve an example protected file url (will be valid ~2min)
final record = await pb.collection('example').getOne('RECORD_ID');
final url = pb.files.getURL(record, record.getStringValue('myPrivateFile'), token: fileToken);   ### Storage options
By default PocketBase stores uploaded files in the pb_data/storage directory on the local file
system. For the majority of cases this is usually the recommended storage option because it is very fast, easy
to work with and backup.
But if you have limited disk space you could switch to an external S3 compatible storage (AWS S3, MinIO,
Wasabi, DigitalOcean Spaces, Vultr Object Storage, etc.). The easiest way to set up the connection
settings is from the Dashboard &gt; Settings &gt; Files storage:

## 16.Web APIs reference - API Collections
page(Number):The page (aka. offset) of the paginated list (default to 1).
perPage(Number):The max returned collections per page (default to 30).
sort(String):Specify the ORDER BY fields. Add - / + (default) in front of the attribute for DESC /
ASC order, e.g.: // DESC by created and ASC by id
?sort=-created,id Supported collection sort fields: @random, id, created,
updated, name, type,
system
filter(String):Filter expression to filter/search the returned collections list, e.g.: ?filter=(name~'abc' && created>'2022-01-01') Supported collection filter fields: id, created, updated,
name, type, system The syntax basically follows the format
OPERAND OPERATOR OPERAND, where: OPERAND - could be any field literal, string (single or double quoted),
number, null, true, false OPERATOR - is one of:
= Equal != NOT equal > Greater than >= Greater than or equal < Less than <= Less than or equal ~ Like/Contains (if not specified auto wraps the right string OPERAND in a "%" for wildcard
match) !~ NOT Like/Contains (if not specified auto wraps the right string OPERAND in a "%" for
wildcard match) ?= Any/At least one of Equal ?!= Any/At least one of NOT equal ?> Any/At least one of Greater than ?>= Any/At least one of Greater than or equal ?< Any/At least one of Less than ?<= Any/At least one of Less than or equal ?~ Any/At least one of Like/Contains (if not specified auto wraps the right string OPERAND in a "%" for wildcard
match) ?!~ Any/At least one of NOT Like/Contains (if not specified auto wraps the right string OPERAND in a "%" for
wildcard match) To group and combine several expressions you can use parenthesis
(...), && (AND) and || (OR) tokens. Single line comments are also supported: // Example comment.
fields(String):Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
skipTotal(Boolean):If it is set the total counts query will be skipped and the response fields
totalItems and totalPages will have -1 value.
This could drastically speed up the search queries when the total counters are not needed or cursor based
pagination is used.
For optimization purposes, it is set by default for the
getFirstListItem()
and
getFullList() SDKs methods.
collectionIdOrName(String):ID or name of the collection to view.
fields(String):Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
fields(String):Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
collectionIdOrName(String):ID or name of the collection to view.
fields(String):Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
collectionIdOrName(String):ID or name of the collection to view.
collectionIdOrName(String):ID or name of the collection to truncate.
Required collections(Array<Collection>):List of collections to import (replace and create).
Optional deleteMissing(Boolean):If true all existing collections and schema fields that are not present in the
imported configuration will be deleted, including their related records
data (default to
false).
# Web APIs reference - API Collections
- **page** (Number): The page (aka. offset) of the paginated list (default to 1).
- **perPage** (Number): The max returned collections per page (default to 30).
- **sort** (String): Specify the ORDER BY fields. Add - / + (default) in front of the attribute for DESC /
ASC order, e.g.: // DESC by created and ASC by id
?sort=-created,id Supported collection sort fields: @random, id, created,
updated, name, type,
system
- **filter** (String): Filter expression to filter/search the returned collections list, e.g.: ?filter=(name~'abc' && created>'2022-01-01') Supported collection filter fields: id, created, updated,
name, type, system The syntax basically follows the format
OPERAND OPERATOR OPERAND, where: OPERAND - could be any field literal, string (single or double quoted),
number, null, true, false OPERATOR - is one of:
= Equal != NOT equal > Greater than >= Greater than or equal < Less than <= Less than or equal ~ Like/Contains (if not specified auto wraps the right string OPERAND in a "%" for wildcard
match) !~ NOT Like/Contains (if not specified auto wraps the right string OPERAND in a "%" for
wildcard match) ?= Any/At least one of Equal ?!= Any/At least one of NOT equal ?> Any/At least one of Greater than ?>= Any/At least one of Greater than or equal ?< Any/At least one of Less than ?<= Any/At least one of Less than or equal ?~ Any/At least one of Like/Contains (if not specified auto wraps the right string OPERAND in a "%" for wildcard
match) ?!~ Any/At least one of NOT Like/Contains (if not specified auto wraps the right string OPERAND in a "%" for
wildcard match) To group and combine several expressions you can use parenthesis
(...), && (AND) and || (OR) tokens. Single line comments are also supported: // Example comment.
- **fields** (String): Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
- **skipTotal** (Boolean): If it is set the total counts query will be skipped and the response fields
totalItems and totalPages will have -1 value.
This could drastically speed up the search queries when the total counters are not needed or cursor based
pagination is used.
For optimization purposes, it is set by default for the
getFirstListItem()
and
getFullList() SDKs methods.
- **collectionIdOrName** (String): ID or name of the collection to view.
- **fields** (String): Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
- **fields** (String): Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
- **collectionIdOrName** (String): ID or name of the collection to view.
- **fields** (String): Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
- **collectionIdOrName** (String): ID or name of the collection to view.
- **collectionIdOrName** (String): ID or name of the collection to truncate.
- **Required collections** (Array<Collection>): List of collections to import (replace and create).
- **Optional deleteMissing** (Boolean): If true all existing collections and schema fields that are not present in the
imported configuration will be deleted, including their related records
data (default to
false).
API Collections    List collections    Returns a paginated Collections list.
Only superusers can perform this action.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
// fetch a paginated collections list
const pageResult = await pb.collections.getList(1, 100, {
filter: 'created >= "2022-01-01 00:00:00"',
// you can also fetch all collections at once via getFullList
const collections = await pb.collections.getFullList({ sort: '-created' });
// or fetch only the first collection that matches the specified filter
const collection = await pb.collections.getFirstListItem('type="auth"');  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
// fetch a paginated collections list
final pageResult = await pb.collections.getList(
page: 1,
perPage: 100,
filter: 'created >= "2022-01-01 00:00:00"',
// you can also fetch all collections at once via getFullList
final collections = await pb.collections.getFullList(sort: '-created');
// or fetch only the first collection that matches the specified filter
final collection = await pb.collections.getFirstListItem('type="auth"');   ###### API details
GET /api/collections Requires `Authorization:TOKEN` Query parameters Param Type Description page Number The page (aka. offset) of the paginated list (default to 1). perPage Number The max returned collections per page (default to 30). sort String Specify the ORDER BY fields.
Add - / + (default) in front of the attribute for DESC /
ASC order, e.g.:
// DESC by created and ASC by id
?sort=-created,id  Supported collection sort fields:  @random, id, created,
updated, name, type,
system
filter String Filter expression to filter/search the returned collections list, e.g.:
`?filter=(name~'abc' &amp;&amp; created>'2022-01-01')`  Supported collection filter fields:  id, created, updated,
name, type, system
The syntax basically follows the format
OPERAND OPERATOR OPERAND, where:
-OPERAND - could be any field literal, string (single or double quoted),
number, null, true, false
-OPERATOR - is one of:
= Equal
-!= NOT equal
-> Greater than
->= Greater than or equal
-&lt; Less than
-&lt;= Less than or equal
-~ Like/Contains (if not specified auto wraps the right string OPERAND in a &quot;%&quot; for wildcard
match)
-!~ NOT Like/Contains (if not specified auto wraps the right string OPERAND in a &quot;%&quot; for
wildcard match)
-?= Any/At least one of Equal
-?!= Any/At least one of NOT equal
-?> Any/At least one of Greater than
-?>= Any/At least one of Greater than or equal
-?&lt; Any/At least one of Less than
-?&lt;= Any/At least one of Less than or equal
-?~ Any/At least one of Like/Contains (if not specified auto wraps the right string OPERAND in a &quot;%&quot; for wildcard
match)
-?!~ Any/At least one of NOT Like/Contains (if not specified auto wraps the right string OPERAND in a &quot;%&quot; for
wildcard match)
To group and combine several expressions you can use parenthesis
(...), &amp;&amp; (AND) and || (OR) tokens.
Single line comments are also supported: // Example comment.
fields String Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name
* targets all keys from the specific depth level.
In addition, the following field modifiers are also supported:
-:excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
skipTotal Boolean If it is set the total counts query will be skipped and the response fields
`totalItems` and `totalPages` will have `-1` value.
This could drastically speed up the search queries when the total counters are not needed or cursor based
pagination is used.
For optimization purposes, it is set by default for the
`getFirstListItem()`
and
`getFullList()` SDKs methods. Responses  {
"page": 1,
"perPage": 2,
"totalItems": 10,
"totalPages": 5,
"items": [
"id": "_pbc_344172009",
"listRule": null,
"viewRule": null,
"createRule": null,
"updateRule": null,
"deleteRule": null,
"name": "users",
"type": "auth",
"fields": [
"autogeneratePattern": "[a-z0-9]{15}",
"hidden": false,
"id": "text3208210256",
"max": 15,
"min": 15,
"name": "id",
"pattern": "^[a-z0-9]+$",
"presentable": false,
"primaryKey": true,
"required": true,
"system": true,
"type": "text"
"cost": 0,
"hidden": true,
"id": "password901924565",
"max": 0,
"min": 8,
"name": "password",
"pattern": "",
"presentable": false,
"required": true,
"system": true,
"type": "password"
"autogeneratePattern": "[a-zA-Z0-9]{50}",
"hidden": true,
"id": "text2504183744",
"max": 60,
"min": 30,
"name": "tokenKey",
"pattern": "",
"presentable": false,
"primaryKey": false,
"required": true,
"system": true,
"type": "text"
"exceptDomains": null,
"hidden": false,
"id": "email3885137012",
"name": "email",
"onlyDomains": null,
"presentable": false,
"required": true,
"system": true,
"type": "email"
"hidden": false,
"id": "bool1547992806",
"name": "emailVisibility",
"presentable": false,
"required": false,
"system": true,
"type": "bool"
"hidden": false,
"id": "bool256245529",
"name": "verified",
"presentable": false,
"required": false,
"system": true,
"type": "bool"
"autogeneratePattern": "",
"hidden": false,
"id": "text1579384326",
"max": 255,
"min": 0,
"name": "name",
"pattern": "",
"presentable": false,
"primaryKey": false,
"required": false,
"system": false,
"type": "text"
"hidden": false,
"id": "file376926767",
"maxSelect": 1,
"maxSize": 0,
"mimeTypes": [
"image/jpeg",
"image/png",
"image/svg+xml",
"image/gif",
"image/webp"
"name": "avatar",
"presentable": false,
"protected": false,
"required": false,
"system": false,
"thumbs": null,
"type": "file"
"hidden": false,
"id": "autodate2990389176",
"name": "created",
"onCreate": true,
"onUpdate": false,
"presentable": false,
"system": false,
"type": "autodate"
"hidden": false,
"id": "autodate3332085495",
"name": "updated",
"onCreate": true,
"onUpdate": true,
"presentable": false,
"system": false,
"type": "autodate"
"indexes": [
"CREATE UNIQUE INDEX `idx_tokenKey__pbc_344172009` ON `users` (`tokenKey`)",
"CREATE UNIQUE INDEX `idx_email__pbc_344172009` ON `users` (`email`) WHERE `email` != ''"
"system": false,
"authRule": "",
"manageRule": null,
"authAlert": {
"enabled": true,
"emailTemplate": {
"subject": "Login from a new location",
"body": "..."
"oauth2": {
"enabled": false,
"mappedFields": {
"id": "",
"name": "name",
"username": "",
"avatarURL": "avatar"
"providers": [
"pkce": null,
"name": "google",
"clientId": "abc",
"authURL": "",
"tokenURL": "",
"userInfoURL": "",
"displayName": "",
"extra": null
"passwordAuth": {
"enabled": true,
"identityFields": [
"email"
"mfa": {
"enabled": false,
"duration": 1800,
"rule": ""
"otp": {
"enabled": false,
"duration": 180,
"length": 8,
"emailTemplate": {
"subject": "OTP for {APP_NAME}",
"body": "..."
"authToken": {
"duration": 604800
"passwordResetToken": {
"duration": 1800
"emailChangeToken": {
"duration": 1800
"verificationToken": {
"duration": 259200
"fileToken": {
"duration": 180
"verificationTemplate": {
"subject": "Verify your {APP_NAME} email",
"body": "..."
"resetPasswordTemplate": {
"subject": "Reset your {APP_NAME} password",
"body": "..."
"confirmEmailChangeTemplate": {
"subject": "Confirm your {APP_NAME} new email address",
"body": "..."
"id": "_pbc_2287844090",
"listRule": null,
"viewRule": null,
"createRule": null,
"updateRule": null,
"deleteRule": null,
"name": "posts",
"type": "base",
"fields": [
"autogeneratePattern": "[a-z0-9]{15}",
"hidden": false,
"id": "text3208210256",
"max": 15,
"min": 15,
"name": "id",
"pattern": "^[a-z0-9]+$",
"presentable": false,
"primaryKey": true,
"required": true,
"system": true,
"type": "text"
"autogeneratePattern": "",
"hidden": false,
"id": "text724990059",
"max": 0,
"min": 0,
"name": "title",
"pattern": "",
"presentable": false,
"primaryKey": false,
"required": false,
"system": false,
"type": "text"
"hidden": false,
"id": "autodate2990389176",
"name": "created",
"onCreate": true,
"onUpdate": false,
"presentable": false,
"system": false,
"type": "autodate"
"hidden": false,
"id": "autodate3332085495",
"name": "updated",
"onCreate": true,
"onUpdate": true,
"presentable": false,
"system": false,
"type": "autodate"
"indexes": [],
"system": false
}  {
"status": 400,
"message": "Something went wrong while processing your request. Invalid filter.",
"data": {}
}  {
"status": 401,
"message": "The request requires valid record authorization token.",
"data": {}
}  {
"status": 403,
"message": "Only superusers can perform this action.",
"data": {}
}      View collection    Returns a single Collection by its ID or name.
Only superusers can perform this action.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
const collection = await pb.collections.getOne('demo');  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
final collection = await pb.collections.getOne('demo');   ###### API details
GET /api/collections/`collectionIdOrName` Requires `Authorization:TOKEN` Path parameters Param Type Description collectionIdOrName String ID or name of the collection to view. Query parameters Param Type Description fields String Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name
* targets all keys from the specific depth level.
In addition, the following field modifiers are also supported:
-:excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
Responses  {
"id": "_pbc_2287844090",
"listRule": null,
"viewRule": null,
"createRule": null,
"updateRule": null,
"deleteRule": null,
"name": "posts",
"type": "base",
"fields": [
"autogeneratePattern": "[a-z0-9]{15}",
"hidden": false,
"id": "text3208210256",
"max": 15,
"min": 15,
"name": "id",
"pattern": "^[a-z0-9]+$",
"presentable": false,
"primaryKey": true,
"required": true,
"system": true,
"type": "text"
"autogeneratePattern": "",
"hidden": false,
"id": "text724990059",
"max": 0,
"min": 0,
"name": "title",
"pattern": "",
"presentable": false,
"primaryKey": false,
"required": false,
"system": false,
"type": "text"
"hidden": false,
"id": "autodate2990389176",
"name": "created",
"onCreate": true,
"onUpdate": false,
"presentable": false,
"system": false,
"type": "autodate"
"hidden": false,
"id": "autodate3332085495",
"name": "updated",
"onCreate": true,
"onUpdate": true,
"presentable": false,
"system": false,
"type": "autodate"
"indexes": [],
"system": false
}  {
"status": 401,
"message": "The request requires valid record authorization token.",
"data": {}
}  {
"status": 403,
"message": "The authorized record is not allowed to perform this action.",
"data": {}
}  {
"status": 404,
"message": "The requested resource wasn't found.",
"data": {}
}      Create collection    Creates a new Collection.
Only superusers can perform this action.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
// create base collection
const base = await pb.collections.create({
name: 'exampleBase',
type: 'base',
fields: [
name: 'title',
type: 'text',
required: true,
min: 10,
name: 'status',
type: 'bool',
// create auth collection
const auth = await pb.collections.create({
name: 'exampleAuth',
type: 'auth',
createRule: 'id = @request.auth.id',
updateRule: 'id = @request.auth.id',
deleteRule: 'id = @request.auth.id',
fields: [
name: 'name',
type: 'text',
passwordAuth: {
enabled: true,
identityFields: ['email']
// create view collection
const view = await pb.collections.create({
name: 'exampleView',
type: 'view',
listRule: '@request.auth.id != ""',
viewRule: null,
// the schema will be autogenerated from the below query
viewQuery: 'SELECT id, name from posts',
});  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
// create base collection
final base = await pb.collections.create(body: {
'name': 'exampleBase',
'type': 'base',
'fields': [
'name': 'title',
'type': 'text',
'required': true,
'min': 10,
'name': 'status',
'type': 'bool',
// create auth collection
final auth = await pb.collections.create(body: {
'name': 'exampleAuth',
'type': 'auth',
'createRule': 'id = @request.auth.id',
'updateRule': 'id = @request.auth.id',
'deleteRule': 'id = @request.auth.id',
'fields': [
'name': 'name',
'type': 'text',
'passwordAuth': {
'enabled': true,
'identityFields': ['email']
// create view collection
final view = await pb.collections.create(body: {
'name': 'exampleView',
'type': 'view',
'listRule': '@request.auth.id != ""',
'viewRule': null,
// the schema will be autogenerated from the below query
'viewQuery': 'SELECT id, name from posts',
});   ###### API details
POST /api/collections Requires `Authorization:TOKEN` Body Parameters Body parameters could be sent as JSON or multipart/form-data.
// 15 characters string to store as collection ID.
// If not set, it will be auto generated.
id (optional): string
// Unique collection name (used as a table name for the records table).
name (required):  string
// Type of the collection.
// If not set, the collection type will be "base" by default.
type (optional): "base" | "view" | "auth"
// List with the collection fields.
// This field is optional and autopopulated for "view" collections based on the viewQuery.
fields (required|optional): Array&lt;Field>
// The collection indexes and unique constraints.
// Note that "view" collections don't support indexes.
indexes (optional): Array&lt;string>
// Marks the collection as "system" to prevent being renamed, deleted or modify its API rules.
system (optional): boolean
// CRUD API rules
listRule (optional):   null|string
viewRule (optional):   null|string
createRule (optional): null|string
updateRule (optional): null|string
deleteRule (optional): null|string
// view options
viewQuery (required):  string
// auth options
// API rule that gives admin-like permissions to allow fully managing the auth record(s),
// e.g. changing the password without requiring to enter the old one, directly updating the
// verified state or email, etc. This rule is executed in addition to the createRule and updateRule.
manageRule (optional): null|string
// API rule that could be used to specify additional record constraints applied after record
// authentication and right before returning the auth token response to the client.
// For example, to allow only verified users you could set it to "verified = true".
// Set it to empty string to allow any Auth collection record to authenticate.
// Set it to null to disallow authentication altogether for the collection.
authRule (optional): null|string
// AuthAlert defines options related to the auth alerts on new device login.
authAlert (optional): {
enabled (optional): boolean
emailTemplate (optional): {
subject (required): string
body (required):    string
// OAuth2 specifies whether OAuth2 auth is enabled for the collection
// and which OAuth2 providers are allowed.
oauth2 (optional): {
enabled (optional): boolean
mappedFields (optional): {
id (optional):        string
name (optional):      string
username (optional):  string
avatarURL (optional): string
providers (optional): [
name (required):         string
clientId (required):     string
clientSecret (required): string
authUrl (optional):      string
tokenUrl (optional):     string
userApiUrl (optional):   string
displayName (optional):  string
pkce (optional):         null|boolean
// PasswordAuth defines options related to the collection password authentication.
passwordAuth (optional): {
enabled (optional):        boolean
identityFields (required): Array&lt;string>
// MFA defines options related to the Multi-factor authentication (MFA).
mfa (optional):{
enabled (optional):  boolean
duration (required): number
rule (optional):     string
// OTP defines options related to the One-time password authentication (OTP).
otp (optional): {
enabled (optional):  boolean
duration (required): number
length (required):   number
emailTemplate (optional): {
subject (required): string
body (required):    string
// Token configurations.
authToken (optional): {
duration (required): number
secret (required):   string
passwordResetToken (optional): {
duration (required): number
secret (required):   string
emailChangeToken (optional): {
duration (required): number
secret (required):   string
verificationToken (optional): {
duration (required): number
secret (required):   string
fileToken (optional): {
duration (required): number
secret (required):   string
// Default email templates.
verificationTemplate (optional): {
subject (required): string
body (required):    string
resetPasswordTemplate (optional): {
subject (required): string
body (required):    string
confirmEmailChangeTemplate (optional): {
subject (required): string
body (required):    string
}  Query parameters Param Type Description fields String Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name
* targets all keys from the specific depth level.
In addition, the following field modifiers are also supported:
-:excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
Responses  {
"id": "_pbc_2287844090",
"listRule": null,
"viewRule": null,
"createRule": null,
"updateRule": null,
"deleteRule": null,
"name": "posts",
"type": "base",
"fields": [
"autogeneratePattern": "[a-z0-9]{15}",
"hidden": false,
"id": "text3208210256",
"max": 15,
"min": 15,
"name": "id",
"pattern": "^[a-z0-9]+$",
"presentable": false,
"primaryKey": true,
"required": true,
"system": true,
"type": "text"
"autogeneratePattern": "",
"hidden": false,
"id": "text724990059",
"max": 0,
"min": 0,
"name": "title",
"pattern": "",
"presentable": false,
"primaryKey": false,
"required": false,
"system": false,
"type": "text"
"hidden": false,
"id": "autodate2990389176",
"name": "created",
"onCreate": true,
"onUpdate": false,
"presentable": false,
"system": false,
"type": "autodate"
"hidden": false,
"id": "autodate3332085495",
"name": "updated",
"onCreate": true,
"onUpdate": true,
"presentable": false,
"system": false,
"type": "autodate"
"indexes": [],
"system": false
}  {
"status": 400,
"message": "An error occurred while submitting the form.",
"data": {
"title": {
"code": "validation_required",
"message": "Missing required value."
}  {
"status": 401,
"message": "The request requires valid record authorization token.",
"data": {}
}  {
"status": 403,
"message": "The authorized record is not allowed to perform this action.",
"data": {}
}      Update collection    Updates a single Collection by its ID or name.
Only superusers can perform this action.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
const collection = await pb.collections.update('demo', {
name: 'new_demo',
listRule: 'created > "2022-01-01 00:00:00"',
});  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
final collection = await pb.collections.update('demo', body: {
'name': 'new_demo',
'listRule': 'created > "2022-01-01 00:00:00"',
});   ###### API details
PATCH /api/collections/`collectionIdOrName` Requires `Authorization:TOKEN` Path parameters Param Type Description collectionIdOrName String ID or name of the collection to view. Body Parameters Body parameters could be sent as JSON or multipart/form-data.
// Unique collection name (used as a table name for the records table).
name (required):  string
// List with the collection fields.
// This field is optional and autopopulated for "view" collections based on the viewQuery.
fields (required|optional): Array&lt;Field>
// The collection indexes and unique constriants.
// Note that "view" collections don't support indexes.
indexes (optional): Array&lt;string>
// Marks the collection as "system" to prevent being renamed, deleted or modify its API rules.
system (optional): boolean
// CRUD API rules
listRule (optional):   null|string
viewRule (optional):   null|string
createRule (optional): null|string
updateRule (optional): null|string
deleteRule (optional): null|string
// view options
viewQuery (required):  string
// auth options
// API rule that gives admin-like permissions to allow fully managing the auth record(s),
// e.g. changing the password without requiring to enter the old one, directly updating the
// verified state or email, etc. This rule is executed in addition to the createRule and updateRule.
manageRule (optional): null|string
// API rule that could be used to specify additional record constraints applied after record
// authentication and right before returning the auth token response to the client.
// For example, to allow only verified users you could set it to "verified = true".
// Set it to empty string to allow any Auth collection record to authenticate.
// Set it to null to disallow authentication altogether for the collection.
authRule (optional): null|string
// AuthAlert defines options related to the auth alerts on new device login.
authAlert (optional): {
enabled (optional): boolean
emailTemplate (optional): {
subject (required): string
body (required):    string
// OAuth2 specifies whether OAuth2 auth is enabled for the collection
// and which OAuth2 providers are allowed.
oauth2 (optional): {
enabled (optional): boolean
mappedFields (optional): {
id (optional):        string
name (optional):      string
username (optional):  string
avatarURL (optional): string
providers (optional): [
name (required):         string
clientId (required):     string
clientSecret (required): string
authUrl (optional):      string
tokenUrl (optional):     string
userApiUrl (optional):   string
displayName (optional):  string
pkce (optional):         null|boolean
// PasswordAuth defines options related to the collection password authentication.
passwordAuth (optional): {
enabled (optional):        boolean
identityFields (required): Array&lt;string>
// MFA defines options related to the Multi-factor authentication (MFA).
mfa (optional):{
enabled (optional):  boolean
duration (required): number
rule (optional):     string
// OTP defines options related to the One-time password authentication (OTP).
otp (optional): {
enabled (optional):  boolean
duration (required): number
length (required):   number
emailTemplate (optional): {
subject (required): string
body (required):    string
// Token configurations.
authToken (optional): {
duration (required): number
secret (required):   string
passwordResetToken (optional): {
duration (required): number
secret (required):   string
emailChangeToken (optional): {
duration (required): number
secret (required):   string
verificationToken (optional): {
duration (required): number
secret (required):   string
fileToken (optional): {
duration (required): number
secret (required):   string
// Default email templates.
verificationTemplate (optional): {
subject (required): string
body (required):    string
resetPasswordTemplate (optional): {
subject (required): string
body (required):    string
confirmEmailChangeTemplate (optional): {
subject (required): string
body (required):    string
}  Query parameters Param Type Description fields String Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name
* targets all keys from the specific depth level.
In addition, the following field modifiers are also supported:
-:excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
Responses  {
"id": "_pbc_2287844090",
"listRule": null,
"viewRule": null,
"createRule": null,
"updateRule": null,
"deleteRule": null,
"name": "posts",
"type": "base",
"fields": [
"autogeneratePattern": "[a-z0-9]{15}",
"hidden": false,
"id": "text3208210256",
"max": 15,
"min": 15,
"name": "id",
"pattern": "^[a-z0-9]+$",
"presentable": false,
"primaryKey": true,
"required": true,
"system": true,
"type": "text"
"autogeneratePattern": "",
"hidden": false,
"id": "text724990059",
"max": 0,
"min": 0,
"name": "title",
"pattern": "",
"presentable": false,
"primaryKey": false,
"required": false,
"system": false,
"type": "text"
"hidden": false,
"id": "autodate2990389176",
"name": "created",
"onCreate": true,
"onUpdate": false,
"presentable": false,
"system": false,
"type": "autodate"
"hidden": false,
"id": "autodate3332085495",
"name": "updated",
"onCreate": true,
"onUpdate": true,
"presentable": false,
"system": false,
"type": "autodate"
"indexes": [],
"system": false
}  {
"status": 400,
"message": "An error occurred while submitting the form.",
"data": {
"email": {
"code": "validation_required",
"message": "Missing required value."
}  {
"status": 401,
"message": "The request requires valid record authorization token.",
"data": {}
}  {
"status": 403,
"message": "The authorized record is not allowed to perform this action.",
"data": {}
}      Delete collection    Deletes a single Collection by its ID or name.
Only superusers can perform this action.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
await pb.collections.delete('demo');  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
await pb.collections.delete('demo');   ###### API details
DELETE /api/collections/`collectionIdOrName` Requires `Authorization:TOKEN` Path parameters Param Type Description collectionIdOrName String ID or name of the collection to view. Responses  `null`  {
"status": 400,
"message": "Failed to delete collection. Make sure that the collection is not referenced by other collections.",
"data": {}
}  {
"status": 401,
"message": "The request requires valid record authorization token.",
"data": {}
}  {
"status": 403,
"message": "The authorized record is not allowed to perform this action.",
"data": {}
}  {
"status": 404,
"message": "The requested resource wasn't found.",
"data": {}
}      Truncate collection    Deletes all the records of a single collection (including their related files and cascade delete
enabled relations).
Only superusers can perform this action.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
await pb.collections.truncate('demo');  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
await pb.collections.truncate('demo');   ###### API details
DELETE /api/collections/`collectionIdOrName`/truncate Requires `Authorization:TOKEN` Path parameters Param Type Description collectionIdOrName String ID or name of the collection to truncate. Responses  `null`  {
"status": 400,
"message": "Failed to truncate collection (most likely due to required cascade delete record references).",
"data": {}
}  {
"status": 401,
"message": "The request requires valid record authorization token.",
"data": {}
}  {
"status": 403,
"message": "The authorized record is not allowed to perform this action.",
"data": {}
}  {
"status": 404,
"message": "The requested resource wasn't found.",
"data": {}
}      Import collections    Bulk imports the provided Collections configuration.
Only superusers can perform this action.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
const importData = [
name: 'collection1',
schema: [
name: 'status',
type: 'bool',
name: 'collection2',
schema: [
name: 'title',
type: 'text',
await pb.collections.import(importData, false);  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
final importData = [
CollectionModel(
name: "collection1",
schema: [
SchemaField(name: "status", type: "bool"),
CollectionModel(
name: "collection2",
schema: [
SchemaField(name: "title", type: "text"),
await pb.collections.import(importData, deleteMissing: false);   ###### API details
PUT /api/collections/import Requires `Authorization:TOKEN` Body Parameters Param Type Description Required collections Array&lt;Collection> List of collections to import (replace and create). Optional deleteMissing Boolean If true all existing collections and schema fields that are not present in the
imported configuration will be deleted, including their related records
data (default to
false). Body parameters could be sent as JSON or
multipart/form-data. Responses  `null`  {
"status": 400,
"message": "An error occurred while submitting the form.",
"data": {
"collections": {
"code": "collections_import_failure",
"message": "Failed to import the collections configuration."
}  {
"status": 401,
"message": "The request requires valid record authorization token.",
"data": {}
}  {
"status": 403,
"message": "The authorized record is not allowed to perform this action.",
"data": {}
}      Scaffolds    Returns an object will all of the collection types and their default fields
(used primarily in the Dashboard UI).
Only superusers can perform this action.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
const scaffolds = await pb.collections.getScaffolds();  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
final scaffolds = await pb.collections.getScaffolds();   ###### API details
GET /api/collections/meta/scaffolds Requires `Authorization:TOKEN` Responses  {
"auth": {
"id": "",
"listRule": null,
"viewRule": null,
"createRule": null,
"updateRule": null,
"deleteRule": null,
"name": "",
"type": "auth",
"fields": [
"autogeneratePattern": "[a-z0-9]{15}",
"hidden": false,
"id": "text3208210256",
"max": 15,
"min": 15,
"name": "id",
"pattern": "^[a-z0-9]+$",
"presentable": false,
"primaryKey": true,
"required": true,
"system": true,
"type": "text"
"cost": 0,
"hidden": true,
"id": "password901924565",
"max": 0,
"min": 8,
"name": "password",
"pattern": "",
"presentable": false,
"required": true,
"system": true,
"type": "password"
"autogeneratePattern": "[a-zA-Z0-9]{50}",
"hidden": true,
"id": "text2504183744",
"max": 60,
"min": 30,
"name": "tokenKey",
"pattern": "",
"presentable": false,
"primaryKey": false,
"required": true,
"system": true,
"type": "text"
"exceptDomains": null,
"hidden": false,
"id": "email3885137012",
"name": "email",
"onlyDomains": null,
"presentable": false,
"required": true,
"system": true,
"type": "email"
"hidden": false,
"id": "bool1547992806",
"name": "emailVisibility",
"presentable": false,
"required": false,
"system": true,
"type": "bool"
"hidden": false,
"id": "bool256245529",
"name": "verified",
"presentable": false,
"required": false,
"system": true,
"type": "bool"
"indexes": [
"CREATE UNIQUE INDEX `idx_tokenKey_hclGvwhtqG` ON `test` (`tokenKey`)",
"CREATE UNIQUE INDEX `idx_email_eyxYyd3gp1` ON `test` (`email`) WHERE `email` != ''"
"created": "",
"updated": "",
"system": false,
"authRule": "",
"manageRule": null,
"authAlert": {
"enabled": true,
"emailTemplate": {
"subject": "Login from a new location",
"body": "..."
"oauth2": {
"providers": [],
"mappedFields": {
"id": "",
"name": "",
"username": "",
"avatarURL": ""
"enabled": false
"passwordAuth": {
"enabled": true,
"identityFields": [
"email"
"mfa": {
"enabled": false,
"duration": 1800,
"rule": ""
"otp": {
"enabled": false,
"duration": 180,
"length": 8,
"emailTemplate": {
"subject": "OTP for {APP_NAME}",
"body": "..."
"authToken": {
"duration": 604800
"passwordResetToken": {
"duration": 1800
"emailChangeToken": {
"duration": 1800
"verificationToken": {
"duration": 259200
"fileToken": {
"duration": 180
"verificationTemplate": {
"subject": "Verify your {APP_NAME} email",
"body": "..."
"resetPasswordTemplate": {
"subject": "Reset your {APP_NAME} password",
"body": "..."
"confirmEmailChangeTemplate": {
"subject": "Confirm your {APP_NAME} new email address",
"body": "..."
"base": {
"id": "",
"listRule": null,
"viewRule": null,
"createRule": null,
"updateRule": null,
"deleteRule": null,
"name": "",
"type": "base",
"fields": [
"autogeneratePattern": "[a-z0-9]{15}",
"hidden": false,
"id": "text3208210256",
"max": 15,
"min": 15,
"name": "id",
"pattern": "^[a-z0-9]+$",
"presentable": false,
"primaryKey": true,
"required": true,
"system": true,
"type": "text"
"indexes": [],
"created": "",
"updated": "",
"system": false
"view": {
"id": "",
"listRule": null,
"viewRule": null,
"createRule": null,
"updateRule": null,
"deleteRule": null,
"name": "",
"type": "view",
"fields": [],
"indexes": [],
"created": "",
"updated": "",
"system": false,
"viewQuery": ""
}  {
"status": 401,
"message": "The request requires valid record authorization token.",
"data": {}
}  {
"status": 403,
"message": "The authorized record is not allowed to perform this action.",
"data": {}
}  {
"status": 404,
"message": "The requested resource wasn't found.",
"data": {}

## 17.Web APIs reference - API Records
POST /api/collections/{collection}/records
collectionIdOrName(String):ID or name of the records' collection.
page(Number):The page (aka. offset) of the paginated list (default to 1).
perPage(Number):The max returned records per page (default to 30).
sort(String):Specify the ORDER BY fields. Add - / + (default) in front of the attribute for DESC /
ASC order, eg.: // DESC by created and ASC by id
?sort=-created,id Supported record sort fields: @random, @rowid, id,
and any other collection field.
filter(String):Filter expression to filter/search the returned records list (in addition to the
collection's listRule), e.g.: ?filter=(title~'abc' && created>'2022-01-01') Supported record filter fields: id, + any field from the collection schema. The syntax basically follows the format
OPERAND OPERATOR OPERAND, where: OPERAND - could be any field literal, string (single or double quoted),
number, null, true, false OPERATOR - is one of:
= Equal != NOT equal > Greater than >= Greater than or equal < Less than <= Less than or equal ~ Like/Contains (if not specified auto wraps the right string OPERAND in a "%" for wildcard
match) !~ NOT Like/Contains (if not specified auto wraps the right string OPERAND in a "%" for
wildcard match) ?= Any/At least one of Equal ?!= Any/At least one of NOT equal ?> Any/At least one of Greater than ?>= Any/At least one of Greater than or equal ?< Any/At least one of Less than ?<= Any/At least one of Less than or equal ?~ Any/At least one of Like/Contains (if not specified auto wraps the right string OPERAND in a "%" for wildcard
match) ?!~ Any/At least one of NOT Like/Contains (if not specified auto wraps the right string OPERAND in a "%" for
wildcard match) To group and combine several expressions you can use parenthesis
(...), && (AND) and || (OR) tokens. Single line comments are also supported: // Example comment.
expand(String):Auto expand record relations. Ex.:
?expand=relField1,relField2.subRelField
Supports up to 6-levels depth nested relations expansion.
The expanded relations will be appended to the record under the
expand property (e.g. "expand": {"relField1": {...}, ...}).
Only the relations to which the request user has permissions to view will be expanded.
fields(String):Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
skipTotal(Boolean):If it is set the total counts query will be skipped and the response fields
totalItems and totalPages will have -1 value.
This could drastically speed up the search queries when the total counters are not needed or cursor based
pagination is used.
For optimization purposes, it is set by default for the
getFirstListItem()
and
getFullList() SDKs methods.
collectionIdOrName(String):ID or name of the record's collection.
recordId(String):ID of the record to view.
expand(String):Auto expand record relations. Ex.:
?expand=relField1,relField2.subRelField
Supports up to 6-levels depth nested relations expansion.
The expanded relations will be appended to the record under the
expand property (e.g. "expand": {"relField1": {...}, ...}).
Only the relations to which the request user has permissions to view will be expanded.
fields(String):Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
collectionIdOrName(String):ID or name of the record's collection.
Optional id(String):15 characters string to store as record ID.
If not set, it will be auto generated.
Required password(String):Auth record password.
Required passwordConfirm(String):Auth record password confirmation.
expand(String):Auto expand record relations. Ex.:
?expand=relField1,relField2.subRelField
Supports up to 6-levels depth nested relations expansion.
The expanded relations will be appended to the record under the
expand property (e.g. "expand": {"relField1": {...}, ...}).
Only the relations to which the request user has permissions to view will be expanded.
fields(String):Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
collectionIdOrName(String):ID or name of the record's collection.
recordId(String):ID of the record to update.
Optional oldPassword*(String):Old auth record password.
This field is required only when changing the record password. Superusers and auth records
with "Manage" access can skip this field.
Optional password(String):New auth record password.
Optional passwordConfirm(String):New auth record password confirmation.
expand(String):Auto expand record relations. Ex.:
?expand=relField1,relField2.subRelField
Supports up to 6-levels depth nested relations expansion.
The expanded relations will be appended to the record under the
expand property (e.g. "expand": {"relField1": {...}, ...}).
Only the relations to which the request user has permissions to view will be expanded.
fields(String):Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
collectionIdOrName(String):ID or name of the record's collection.
recordId(String):ID of the record to delete.
Required requests(Array<Request> - List of the requests to process.
The supported batch request actions are: record create - POST /api/collections/{collection}/records record update -
PATCH /api/collections/{collection}/records/{id} record upsert - PUT /api/collections/{collection}/records (the body must have id field) record delete -
DELETE /api/collections/{collection}/records/{id} Each batch Request element have the following properties: url path (could include query parameters) method (GET, POST, PUT, PATCH, DELETE) headers (custom per-request Authorization header is not supported at the moment,
aka. all batch requests have the same auth state) body NB! When the batch request is send as
multipart/form-data, the regular batch action fields are expected to be
submitted as serialized json under the @jsonPayload field and file keys
need to follow the pattern requests.N.fileField or
requests[N].fileField (this is usually handled transparently by the SDKs when their specific object
notation is used)
If you don't use the SDKs or prefer manually to construct the FormData
body, then it could look something like:
const formData = new FormData();
formData.append("@jsonPayload", JSON.stringify({
requests: [
method: "POST",
url: "/api/collections/example/records?expand=user",
body: { title: "test1" },
},
method: "PATCH",
url: "/api/collections/example/records/RECORD_ID",
body: { title: "test2" },
},
method: "DELETE",
url: "/api/collections/example/records/RECORD_ID",
},
}))
// file for the first request
formData.append("requests.0.document", new File(...))
// file for the second request
formData.append("requests.1.document", new File(...))):
collectionIdOrName(String):ID or name of the auth collection.
fields(String):Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
collectionIdOrName(String):ID or name of the auth collection.
Required identity(String):Auth record username or email address.
Required password(String):Auth record password.
Optional identityField(String):A specific identity field to use (by default fallbacks to the first matching one).
expand(String):Auto expand record relations. Ex.:
?expand=relField1,relField2.subRelField
Supports up to 6-levels depth nested relations expansion.
The expanded relations will be appended to the record under the
expand property (e.g. "expand": {"relField1": {...}, ...}).
Only the relations to which the request user has permissions to view will be expanded.
fields(String):Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,record.expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,record.description:excerpt(200,true)
collectionIdOrName(String):ID or name of the auth collection.
Required provider(String):The name of the OAuth2 client provider (e.g. "google").
Required code(String):The authorization code returned from the initial request.
Required codeVerifier(String):The code verifier sent with the initial request as part of the code_challenge.
Required redirectUrl(String):The redirect url sent with the initial request.
Optional createData(Object):Optional data that will be used when creating the auth record on OAuth2 sign-up. The created auth record must comply with the same requirements and validations in the
regular create action.
The data can only be in json, aka. multipart/form-data and
files upload currently are not supported during OAuth2 sign-ups.
expand(String):Auto expand record relations. Ex.:
?expand=relField1,relField2.subRelField
Supports up to 6-levels depth nested relations expansion.
The expanded relations will be appended to the record under the
expand property (e.g. "expand": {"relField1": {...}, ...}).
Only the relations to which the request user has permissions to view will be expanded.
fields(String):Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,record.expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,record.description:excerpt(200,true)
collectionIdOrName(String):ID or name of the auth collection.
Required email(String):The auth record email address to send the OTP request (if exists).
collectionIdOrName(String):ID or name of the auth collection.
Required otpId(String):The id of the OTP request.
Required password(String):The one-time password.
expand(String):Auto expand record relations. Ex.:
?expand=relField1,relField2.subRelField
Supports up to 6-levels depth nested relations expansion.
The expanded relations will be appended to the record under the
expand property (e.g. "expand": {"relField1": {...}, ...}).
Only the relations to which the request user has permissions to view will be expanded.
fields(String):Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,record.expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,record.description:excerpt(200,true)
collectionIdOrName(String):ID or name of the auth collection.
expand(String):Auto expand record relations. Ex.:
?expand=relField1,relField2.subRelField
Supports up to 6-levels depth nested relations expansion.
The expanded relations will be appended to the record under the
expand property (e.g. "expand": {"relField1": {...}, ...}).
Only the relations to which the request user has permissions to view will be expanded.
fields(String):Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,record.expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,record.description:excerpt(200,true)
Required email(String):The auth record email address to send the verification request (if exists).
Required token(String):The token from the verification request email.
Required email(String):The auth record email address to send the password reset request (if exists).
Required token(String):The token from the password reset request email.
Required password(String):The new password to set.
Required passwordConfirm(String):The new password confirmation.
Required newEmail(String):The new email address to send the change email request.
Required token(String):The token from the change email request email.
Required password(String):The account password to confirm the email change.
collectionIdOrName(String):ID or name of the auth collection.
id(String):ID of the auth record to impersonate.
Optional duration(Number):Optional custom JWT duration for the exp claim (in seconds).
If not set or 0, it fallbacks to the default collection auth token duration option.
expand(String):Auto expand record relations. Ex.:
?expand=relField1,relField2.subRelField
Supports up to 6-levels depth nested relations expansion.
The expanded relations will be appended to the record under the
expand property (e.g. "expand": {"relField1": {...}, ...}).
Only the relations to which the request user has permissions to view will be expanded.
fields(String):Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,record.expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,record.description:excerpt(200,true)
# Web APIs reference - API Records
`POST /api/collections/{collection}/records`
- **collectionIdOrName** (String): ID or name of the records' collection.
- **page** (Number): The page (aka. offset) of the paginated list (default to 1).
- **perPage** (Number): The max returned records per page (default to 30).
- **sort** (String): Specify the ORDER BY fields. Add - / + (default) in front of the attribute for DESC /
ASC order, eg.: // DESC by created and ASC by id
?sort=-created,id Supported record sort fields: @random, @rowid, id,
and any other collection field.
- **filter** (String): Filter expression to filter/search the returned records list (in addition to the
collection's listRule), e.g.: ?filter=(title~'abc' && created>'2022-01-01') Supported record filter fields: id, + any field from the collection schema. The syntax basically follows the format
OPERAND OPERATOR OPERAND, where: OPERAND - could be any field literal, string (single or double quoted),
number, null, true, false OPERATOR - is one of:
= Equal != NOT equal > Greater than >= Greater than or equal < Less than <= Less than or equal ~ Like/Contains (if not specified auto wraps the right string OPERAND in a "%" for wildcard
match) !~ NOT Like/Contains (if not specified auto wraps the right string OPERAND in a "%" for
wildcard match) ?= Any/At least one of Equal ?!= Any/At least one of NOT equal ?> Any/At least one of Greater than ?>= Any/At least one of Greater than or equal ?< Any/At least one of Less than ?<= Any/At least one of Less than or equal ?~ Any/At least one of Like/Contains (if not specified auto wraps the right string OPERAND in a "%" for wildcard
match) ?!~ Any/At least one of NOT Like/Contains (if not specified auto wraps the right string OPERAND in a "%" for
wildcard match) To group and combine several expressions you can use parenthesis
(...), && (AND) and || (OR) tokens. Single line comments are also supported: // Example comment.
- **expand** (String): Auto expand record relations. Ex.:
?expand=relField1,relField2.subRelField
Supports up to 6-levels depth nested relations expansion.
The expanded relations will be appended to the record under the
expand property (e.g. "expand": {"relField1": {...}, ...}).
Only the relations to which the request user has permissions to view will be expanded.
- **fields** (String): Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
- **skipTotal** (Boolean): If it is set the total counts query will be skipped and the response fields
totalItems and totalPages will have -1 value.
This could drastically speed up the search queries when the total counters are not needed or cursor based
pagination is used.
For optimization purposes, it is set by default for the
getFirstListItem()
and
getFullList() SDKs methods.
- **collectionIdOrName** (String): ID or name of the record's collection.
- **recordId** (String): ID of the record to view.
- **expand** (String): Auto expand record relations. Ex.:
?expand=relField1,relField2.subRelField
Supports up to 6-levels depth nested relations expansion.
The expanded relations will be appended to the record under the
expand property (e.g. "expand": {"relField1": {...}, ...}).
Only the relations to which the request user has permissions to view will be expanded.
- **fields** (String): Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
- **collectionIdOrName** (String): ID or name of the record's collection.
- **Optional id** (String): 15 characters string to store as record ID.
If not set, it will be auto generated.
- **Required password** (String): Auth record password.
- **Required passwordConfirm** (String): Auth record password confirmation.
- **expand** (String): Auto expand record relations. Ex.:
?expand=relField1,relField2.subRelField
Supports up to 6-levels depth nested relations expansion.
The expanded relations will be appended to the record under the
expand property (e.g. "expand": {"relField1": {...}, ...}).
Only the relations to which the request user has permissions to view will be expanded.
- **fields** (String): Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
- **collectionIdOrName** (String): ID or name of the record's collection.
- **recordId** (String): ID of the record to update.
- **Optional oldPassword** (String) (required): Old auth record password.
This field is required only when changing the record password. Superusers and auth records
with "Manage" access can skip this field.
- **Optional password** (String): New auth record password.
- **Optional passwordConfirm** (String): New auth record password confirmation.
- **expand** (String): Auto expand record relations. Ex.:
?expand=relField1,relField2.subRelField
Supports up to 6-levels depth nested relations expansion.
The expanded relations will be appended to the record under the
expand property (e.g. "expand": {"relField1": {...}, ...}).
Only the relations to which the request user has permissions to view will be expanded.
- **fields** (String): Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
- **collectionIdOrName** (String): ID or name of the record's collection.
- **recordId** (String): ID of the record to delete.
- **Required requests** (Array<Request> - List of the requests to process.
The supported batch request actions are: record create - POST /api/collections/{collection}/records record update -
PATCH /api/collections/{collection}/records/{id} record upsert - PUT /api/collections/{collection}/records (the body must have id field) record delete -
DELETE /api/collections/{collection}/records/{id} Each batch Request element have the following properties: url path (could include query parameters) method (GET, POST, PUT, PATCH, DELETE) headers (custom per-request Authorization header is not supported at the moment,
aka. all batch requests have the same auth state) body NB! When the batch request is send as
multipart/form-data, the regular batch action fields are expected to be
submitted as serialized json under the @jsonPayload field and file keys
need to follow the pattern requests.N.fileField or
requests[N].fileField (this is usually handled transparently by the SDKs when their specific object
notation is used)
If you don't use the SDKs or prefer manually to construct the FormData
body, then it could look something like:
const formData = new FormData();
formData.append("@jsonPayload", JSON.stringify({
requests: [
method: "POST",
url: "/api/collections/example/records?expand=user",
body: { title: "test1" },
method: "PATCH",
url: "/api/collections/example/records/RECORD_ID",
body: { title: "test2" },
method: "DELETE",
url: "/api/collections/example/records/RECORD_ID",
}))
// file for the first request
formData.append("requests.0.document", new File(...))
// file for the second request
formData.append("requests.1.document", new File(...))):
- **collectionIdOrName** (String): ID or name of the auth collection.
- **fields** (String): Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
- **collectionIdOrName** (String): ID or name of the auth collection.
- **Required identity** (String): Auth record username or email address.
- **Required password** (String): Auth record password.
- **Optional identityField** (String): A specific identity field to use (by default fallbacks to the first matching one).
- **expand** (String): Auto expand record relations. Ex.:
?expand=relField1,relField2.subRelField
Supports up to 6-levels depth nested relations expansion.
The expanded relations will be appended to the record under the
expand property (e.g. "expand": {"relField1": {...}, ...}).
Only the relations to which the request user has permissions to view will be expanded.
- **fields** (String): Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,record.expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,record.description:excerpt(200,true)
- **collectionIdOrName** (String): ID or name of the auth collection.
- **Required provider** (String): The name of the OAuth2 client provider (e.g. "google").
- **Required code** (String): The authorization code returned from the initial request.
- **Required codeVerifier** (String): The code verifier sent with the initial request as part of the code_challenge.
- **Required redirectUrl** (String): The redirect url sent with the initial request.
- **Optional createData** (Object): Optional data that will be used when creating the auth record on OAuth2 sign-up. The created auth record must comply with the same requirements and validations in the
regular create action.
The data can only be in json, aka. multipart/form-data and
files upload currently are not supported during OAuth2 sign-ups.
- **expand** (String): Auto expand record relations. Ex.:
?expand=relField1,relField2.subRelField
Supports up to 6-levels depth nested relations expansion.
The expanded relations will be appended to the record under the
expand property (e.g. "expand": {"relField1": {...}, ...}).
Only the relations to which the request user has permissions to view will be expanded.
- **fields** (String): Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,record.expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,record.description:excerpt(200,true)
- **collectionIdOrName** (String): ID or name of the auth collection.
- **Required email** (String): The auth record email address to send the OTP request (if exists).
- **collectionIdOrName** (String): ID or name of the auth collection.
- **Required otpId** (String): The id of the OTP request.
- **Required password** (String): The one-time password.
- **expand** (String): Auto expand record relations. Ex.:
?expand=relField1,relField2.subRelField
Supports up to 6-levels depth nested relations expansion.
The expanded relations will be appended to the record under the
expand property (e.g. "expand": {"relField1": {...}, ...}).
Only the relations to which the request user has permissions to view will be expanded.
- **fields** (String): Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,record.expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,record.description:excerpt(200,true)
- **collectionIdOrName** (String): ID or name of the auth collection.
- **expand** (String): Auto expand record relations. Ex.:
?expand=relField1,relField2.subRelField
Supports up to 6-levels depth nested relations expansion.
The expanded relations will be appended to the record under the
expand property (e.g. "expand": {"relField1": {...}, ...}).
Only the relations to which the request user has permissions to view will be expanded.
- **fields** (String): Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,record.expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,record.description:excerpt(200,true)
- **Required email** (String): The auth record email address to send the verification request (if exists).
- **Required token** (String): The token from the verification request email.
- **Required email** (String): The auth record email address to send the password reset request (if exists).
- **Required token** (String): The token from the password reset request email.
- **Required password** (String): The new password to set.
- **Required passwordConfirm** (String): The new password confirmation.
- **Required newEmail** (String): The new email address to send the change email request.
- **Required token** (String): The token from the change email request email.
- **Required password** (String): The account password to confirm the email change.
- **collectionIdOrName** (String): ID or name of the auth collection.
- **id** (String): ID of the auth record to impersonate.
- **Optional duration** (Number): Optional custom JWT duration for the exp claim (in seconds).
If not set or 0, it fallbacks to the default collection auth token duration option.
- **expand** (String): Auto expand record relations. Ex.:
?expand=relField1,relField2.subRelField
Supports up to 6-levels depth nested relations expansion.
The expanded relations will be appended to the record under the
expand property (e.g. "expand": {"relField1": {...}, ...}).
Only the relations to which the request user has permissions to view will be expanded.
- **fields** (String): Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,record.expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,record.description:excerpt(200,true)
API Records ### CRUD actions
List/Search records    Returns a paginated records list, supporting sorting and filtering.
Depending on the collection&#39;s listRule value, the access to this action may or may not
have been restricted.
You could find individual generated records API documentation in the &quot;Dashboard &gt; Collections
&gt; API Preview&quot;.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
// fetch a paginated records list
const resultList = await pb.collection('posts').getList(1, 50, {
filter: 'created >= "2022-01-01 00:00:00" &amp;&amp; someField1 != someField2',
// you can also fetch all records at once via getFullList
const records = await pb.collection('posts').getFullList({
sort: '-created',
// or fetch only the first record that matches the specified filter
const record = await pb.collection('posts').getFirstListItem('someField="test"', {
expand: 'relField1,relField2.subRelField',
});  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
// fetch a paginated records list
final resultList = await pb.collection('posts').getList(
page: 1,
perPage: 50,
filter: 'created >= "2022-01-01 00:00:00" &amp;&amp; someField1 != someField2',
// you can also fetch all records at once via getFullList
final records = await pb.collection('posts').getFullList(sort: '-created');
// or fetch only the first record that matches the specified filter
final record = await pb.collection('posts').getFirstListItem(
'someField="test"',
expand: 'relField1,relField2.subRelField',
);   ###### API details
GET /api/collections/`collectionIdOrName`/records Path parameters Param Type Description collectionIdOrName String ID or name of the records&#39; collection. Query parameters Param Type Description page Number The page (aka. offset) of the paginated list (default to 1). perPage Number The max returned records per page (default to 30). sort String Specify the ORDER BY fields.
Add - / + (default) in front of the attribute for DESC /
ASC order, eg.:
// DESC by created and ASC by id
?sort=-created,id  Supported record sort fields:  @random, @rowid, id,
and any other collection field.
filter String Filter expression to filter/search the returned records list (in addition to the
collection&#39;s listRule), e.g.:
`?filter=(title~'abc' &amp;&amp; created>'2022-01-01')`  Supported record filter fields:  id, + any field from the collection schema.
The syntax basically follows the format
OPERAND OPERATOR OPERAND, where:
-OPERAND - could be any field literal, string (single or double quoted),
number, null, true, false
-OPERATOR - is one of:
= Equal
-!= NOT equal
-> Greater than
->= Greater than or equal
-&lt; Less than
-&lt;= Less than or equal
-~ Like/Contains (if not specified auto wraps the right string OPERAND in a &quot;%&quot; for wildcard
match)
-!~ NOT Like/Contains (if not specified auto wraps the right string OPERAND in a &quot;%&quot; for
wildcard match)
-?= Any/At least one of Equal
-?!= Any/At least one of NOT equal
-?> Any/At least one of Greater than
-?>= Any/At least one of Greater than or equal
-?&lt; Any/At least one of Less than
-?&lt;= Any/At least one of Less than or equal
-?~ Any/At least one of Like/Contains (if not specified auto wraps the right string OPERAND in a &quot;%&quot; for wildcard
match)
-?!~ Any/At least one of NOT Like/Contains (if not specified auto wraps the right string OPERAND in a &quot;%&quot; for
wildcard match)
To group and combine several expressions you can use parenthesis
(...), &amp;&amp; (AND) and || (OR) tokens.
Single line comments are also supported: // Example comment.
expand String Auto expand record relations. Ex.:
`?expand=relField1,relField2.subRelField`
Supports up to 6-levels depth nested relations expansion.
The expanded relations will be appended to the record under the
`expand` property (e.g. `"expand": {"relField1": {...}, ...}`).
Only the relations to which the request user has permissions to view will be expanded. fields String Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name
* targets all keys from the specific depth level.
In addition, the following field modifiers are also supported:
-:excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
skipTotal Boolean If it is set the total counts query will be skipped and the response fields
`totalItems` and `totalPages` will have `-1` value.
This could drastically speed up the search queries when the total counters are not needed or cursor based
pagination is used.
For optimization purposes, it is set by default for the
`getFirstListItem()`
and
`getFullList()` SDKs methods. Responses  {
"page": 1,
"perPage": 100,
"totalItems": 2,
"totalPages": 1,
"items": [
"id": "ae40239d2bc4477",
"collectionId": "a98f514eb05f454",
"collectionName": "posts",
"updated": "2022-06-25 11:03:50.052",
"created": "2022-06-25 11:03:35.163",
"title": "test1"
"id": "d08dfc4f4d84419",
"collectionId": "a98f514eb05f454",
"collectionName": "posts",
"updated": "2022-06-25 11:03:45.876",
"created": "2022-06-25 11:03:45.876",
"title": "test2"
}  {
"status": 400,
"message": "Something went wrong while processing your request. Invalid filter.",
"data": {}
}  {
"status": 403,
"message": "Only superusers can filter by '@collection.*'",
"data": {}
}      View record    Returns a single collection record by its ID.
Depending on the collection&#39;s viewRule value, the access to this action may or may not
have been restricted.
You could find individual generated records API documentation in the &quot;Dashboard &gt; Collections
&gt; API Preview&quot;.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
const record1 = await pb.collection('posts').getOne('RECORD_ID', {
expand: 'relField1,relField2.subRelField',
});  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
final record1 = await pb.collection('posts').getOne('RECORD_ID',
expand: 'relField1,relField2.subRelField',
);   ###### API details
GET /api/collections/`collectionIdOrName`/records/`recordId` Path parameters Param Type Description collectionIdOrName String ID or name of the record&#39;s collection. recordId String ID of the record to view. Query parameters Param Type Description expand String Auto expand record relations. Ex.:
`?expand=relField1,relField2.subRelField`
Supports up to 6-levels depth nested relations expansion.
The expanded relations will be appended to the record under the
`expand` property (e.g. `"expand": {"relField1": {...}, ...}`).
Only the relations to which the request user has permissions to view will be expanded. fields String Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name
* targets all keys from the specific depth level.
In addition, the following field modifiers are also supported:
-:excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
Responses  {
"id": "ae40239d2bc4477",
"collectionId": "a98f514eb05f454",
"collectionName": "posts",
"updated": "2022-06-25 11:03:50.052",
"created": "2022-06-25 11:03:35.163",
"title": "test1"
}  {
"status": 403,
"message": "Only superusers can perform this action.",
"data": {}
}  {
"status": 404,
"message": "The requested resource wasn't found.",
"data": {}
}      Create record    Creates a new collection Record.
Depending on the collection&#39;s createRule value, the access to this action may or may not
have been restricted.
You could find individual generated records API documentation from the Dashboard.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
const record = await pb.collection('demo').create({
});  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
final record = await pb.collection('demo').create(body: {
});   ###### API details
POST /api/collections/`collectionIdOrName`/records Path parameters Param Type Description collectionIdOrName String ID or name of the record&#39;s collection. Body Parameters Param Type Description Optional id String 15 characters string to store as record ID.
If not set, it will be auto generated. Schema fields Any field from the collection&#39;s schema. Additional auth record fields Required password String Auth record password. Required passwordConfirm String Auth record password confirmation. Body parameters could be sent as JSON or
multipart/form-data.
File upload is supported only through multipart/form-data. Query parameters Param Type Description expand String Auto expand record relations. Ex.:
`?expand=relField1,relField2.subRelField`
Supports up to 6-levels depth nested relations expansion.
The expanded relations will be appended to the record under the
`expand` property (e.g. `"expand": {"relField1": {...}, ...}`).
Only the relations to which the request user has permissions to view will be expanded. fields String Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name
* targets all keys from the specific depth level.
In addition, the following field modifiers are also supported:
-:excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
Responses  {
"collectionId": "a98f514eb05f454",
"collectionName": "demo",
"id": "ae40239d2bc4477",
"updated": "2022-06-25 11:03:50.052",
"created": "2022-06-25 11:03:35.163",
}  {
"status": 400,
"message": "Failed to create record.",
"data": {
"title": {
"code": "validation_required",
"message": "Missing required value."
}  {
"status": 403,
"message": "Only superusers can perform this action.",
"data": {}
}  {
"status": 404,
"message": "The requested resource wasn't found. Missing collection context.",
"data": {}
}      Update record    Updates an existing collection Record.
Depending on the collection&#39;s updateRule value, the access to this action may or may not
have been restricted.
You could find individual generated records API documentation from the Dashboard.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
const record = await pb.collection('demo').update('YOUR_RECORD_ID', {
});  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
final record = await pb.collection('demo').update('YOUR_RECORD_ID', body: {
});   ###### API details
PATCH /api/collections/`collectionIdOrName`/records/`recordId` Path parameters Param Type Description collectionIdOrName String ID or name of the record&#39;s collection. recordId String ID of the record to update. Body Parameters Param Type Description Schema fields Any field from the collection&#39;s schema. Additional auth record fields Optional oldPassword String Old auth record password.
This field is required only when changing the record password. Superusers and auth records
with &quot;Manage&quot; access can skip this field. Optional password String New auth record password. Optional passwordConfirm String New auth record password confirmation. Body parameters could be sent as JSON or
multipart/form-data.
File upload is supported only through multipart/form-data. Query parameters Param Type Description expand String Auto expand record relations. Ex.:
`?expand=relField1,relField2.subRelField`
Supports up to 6-levels depth nested relations expansion.
The expanded relations will be appended to the record under the
`expand` property (e.g. `"expand": {"relField1": {...}, ...}`).
Only the relations to which the request user has permissions to view will be expanded. fields String Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name
* targets all keys from the specific depth level.
In addition, the following field modifiers are also supported:
-:excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
Responses  {
"collectionId": "a98f514eb05f454",
"collectionName": "demo",
"id": "ae40239d2bc4477",
"updated": "2022-06-25 11:03:50.052",
"created": "2022-06-25 11:03:35.163",
}  {
"status": 400,
"message": "Failed to create record.",
"data": {
"title": {
"code": "validation_required",
"message": "Missing required value."
}  {
"status": 403,
"message": "Only superusers can perform this action.",
"data": {}
}  {
"status": 404,
"message": "The requested resource wasn't found. Missing collection context.",
"data": {}
}      Delete record    Deletes a single collection Record by its ID.
Depending on the collection&#39;s deleteRule value, the access to this action may or may not
have been restricted.
You could find individual generated records API documentation from the Dashboard.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
await pb.collection('demo').delete('YOUR_RECORD_ID');  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
await pb.collection('demo').delete('YOUR_RECORD_ID');   ###### API details
DELETE /api/collections/`collectionIdOrName`/records/`recordId` Path parameters Param Type Description collectionIdOrName String ID or name of the record&#39;s collection. recordId String ID of the record to delete. Responses  `null`  {
"status": 400,
"message": "Failed to delete record. Make sure that the record is not part of a required relation reference.",
"data": {}
}  {
"status": 403,
"message": "Only superusers can perform this action.",
"data": {}
}  {
"status": 404,
"message": "The requested resource wasn't found.",
"data": {}
}      Batch create/update/upsert/delete records    Batch and transactional create/update/upsert/delete of multiple records in a single request.
The batch Web API need to be explicitly enabled and configured from the
Dashboard &gt; Settings &gt; Application.
Because this endpoint processes the requests in a single read&amp;write transaction, other queries
may queue up and it could degrade the performance of your application if not used with proper
care and configuration
(some recommendations: prefer using the smallest possible max processing time and body
size limits; avoid large file uploads over slow S3 networks and custom hooks that
communicate with slow external APIs).
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
const batch = pb.createBatch();
batch.collection('example1').create({ ... });
batch.collection('example2').update('RECORD_ID', { ... });
batch.collection('example3').delete('RECORD_ID');
batch.collection('example4').upsert({ ... });
const result = await batch.send();  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
final batch = pb.createBatch();
batch.collection('example1').create(body: { ... });
batch.collection('example2').update('RECORD_ID', body: { ... });
batch.collection('example3').delete('RECORD_ID');
batch.collection('example4').upsert(body: { ... });
final result = await batch.send();   ###### API details
POST /api/batch Body Parameters Body parameters could be sent as application/json or multipart/form-data.
File upload is supported only via multipart/form-data (see below for more details).
Param Description Required requests Array&lt;Request> - List of the requests to process.
The supported batch request actions are:
-record create - POST /api/collections/{collection}/records
-record update -
PATCH /api/collections/{collection}/records/{id}
-record upsert - PUT /api/collections/{collection}/records  (the body must have id field)
-record delete -
DELETE /api/collections/{collection}/records/{id}
Each batch Request element have the following properties:
-url path (could include query parameters)
-method (GET, POST, PUT, PATCH, DELETE)
-headers  (custom per-request Authorization header is not supported at the moment,
aka. all batch requests have the same auth state)
-body
NB! When the batch request is send as
multipart/form-data, the regular batch action fields are expected to be
submitted as serialized json under the @jsonPayload field and file keys
need to follow the pattern requests.N.fileField or
requests[N].fileField (this is usually handled transparently by the SDKs when their specific object
notation is used)
If you don&#39;t use the SDKs or prefer manually to construct the FormData
body, then it could look something like:
const formData = new FormData();
formData.append("@jsonPayload", JSON.stringify({
requests: [
method: "POST",
url: "/api/collections/example/records?expand=user",
body: { title: "test1" },
method: "PATCH",
url: "/api/collections/example/records/RECORD_ID",
body: { title: "test2" },
method: "DELETE",
url: "/api/collections/example/records/RECORD_ID",
// file for the first request
formData.append("requests.0.document", new File(...))
// file for the second request
formData.append("requests.1.document", new File(...))
Responses  [
"status": 200,
"body": {
"collectionId": "a98f514eb05f454",
"collectionName": "demo",
"id": "ae40239d2bc4477",
"updated": "2022-06-25 11:03:50.052",
"created": "2022-06-25 11:03:35.163",
"title": "test1",
"document": "file_a98f51.txt"
"status": 200,
"body": {
"collectionId": "a98f514eb05f454",
"collectionName": "demo",
"id": "31y1gc447bc9602",
"updated": "2022-06-25 11:03:50.052",
"created": "2022-06-25 11:03:35.163",
"title": "test2",
"document": "file_f514eb0.txt"
]  {
"status": 400,
"message": "Batch transaction failed.",
"data": {
"requests": {
"1": {
"code": "batch_request_failed",
"message": "Batch request failed.",
"response": {
"status": 400,
"message": "Failed to create record.",
"data": {
"title": {
"code": "validation_min_text_constraint",
"message": "Must be at least 3 character(s).",
"params": { "min": 3 }
}  {
"status": 403,
"message": "Batch requests are not allowed.",
"data": {}
}   ### Auth record actions
List auth methods    Returns a public list with the allowed collection authentication methods.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
const result = await pb.collection('users').listAuthMethods();  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
final result = await pb.collection('users').listAuthMethods();   ###### API details
GET /api/collections/`collectionIdOrName`/auth-methods Path parameters Param Type Description collectionIdOrName String ID or name of the auth collection. Query parameters Param Type Description fields String Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name
* targets all keys from the specific depth level.
In addition, the following field modifiers are also supported:
-:excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
Responses  {
"password": {
"enabled": true,
"identityFields": ["email"]
"oauth2": {
"enabled": true,
"providers": [
"name": "github",
"displayName": "GitHub",
"state": "nT7SLxzXKAVMeRQJtxSYj9kvnJAvGk",
"authURL": "https://github.com/login/oauth/authorize?client_id=test&amp;code_challenge=fcf8WAhNI6uCLJYgJubLyWXHvfs8xghoLe3zksBvxjE&amp;code_challenge_method=S256&amp;response_type=code&amp;scope=read%3Auser+user%3Aemail&amp;state=nT7SLxzXKAVMeRQJtxSYj9kvnJAvGk&amp;redirect_uri=",
"codeVerifier": "PwBG5OKR2IyQ7siLrrcgWHFwLLLAeUrz7PS1nY4AneG",
"codeChallenge": "fcf8WAhNI6uCLJYgJubLyWXHvfs8xghoLe3zksBvxjE",
"codeChallengeMethod": "S256"
"mfa": {
"enabled": false,
"duration": 0
"otp": {
"enabled": false,
"duration": 0
}      Auth with password    Authenticate a single auth record by combination of a password and a unique identity field (e.g.
email).
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
const authData = await pb.collection('users').authWithPassword(
'YOUR_USERNAME_OR_EMAIL',
'YOUR_PASSWORD',
// after the above you can also access the auth data from the authStore
console.log(pb.authStore.isValid);
console.log(pb.authStore.token);
console.log(pb.authStore.record.id);
// "logout" the last authenticated record
pb.authStore.clear();  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
final authData = await pb.collection('users').authWithPassword(
'YOUR_USERNAME_OR_EMAIL',
'YOUR_PASSWORD',
// after the above you can also access the auth data from the authStore
print(pb.authStore.isValid);
print(pb.authStore.token);
print(pb.authStore.record.id);
// "logout" the last authenticated record
pb.authStore.clear();   ###### API details
POST /api/collections/`collectionIdOrName`/auth-with-password Path parameters Param Type Description collectionIdOrName String ID or name of the auth collection. Body Parameters Param Type Description Required identity String Auth record username or email address. Required password String Auth record password. Optional identityField String A specific identity field to use (by default fallbacks to the first matching one). Body parameters could be sent as JSON or
multipart/form-data. Query parameters Param Type Description expand String Auto expand record relations. Ex.:
`?expand=relField1,relField2.subRelField`
Supports up to 6-levels depth nested relations expansion.
The expanded relations will be appended to the record under the
`expand` property (e.g. `"expand": {"relField1": {...}, ...}`).
Only the relations to which the request user has permissions to view will be expanded. fields String Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,record.expand.relField.name
* targets all keys from the specific depth level.
In addition, the following field modifiers are also supported:
-:excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,record.description:excerpt(200,true)
Responses  {
"token": "eyJhbGciOiJIUzI1NiJ9.eyJpZCI6IjRxMXhsY2xtZmxva3UzMyIsInR5cGUiOiJhdXRoUmVjb3JkIiwiY29sbGVjdGlvbklkIjoiX3BiX3VzZXJzX2F1dGhfIiwiZXhwIjoyMjA4OTg1MjYxfQ.UwD8JvkbQtXpymT09d7J6fdA0aP9g4FJ1GPh_ggEkzc",
"record": {
"id": "8171022dc95a4ed",
"collectionId": "d2972397d45614e",
"collectionName": "users",
"created": "2022-06-24 06:24:18.434Z",
"updated": "2022-06-24 06:24:18.889Z",
"verified": false,
"emailVisibility": true,
"someCustomField": "example 123"
}  {
"status": 400,
"message": "An error occurred while submitting the form.",
"data": {
"password": {
"code": "validation_required",
"message": "Missing required value."
}      Auth with OAuth2    Authenticate with an OAuth2 provider and returns a new auth token and record data.
This action usually should be called right after the provider login page redirect.
You could also check the
OAuth2 web integration example.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
const authData = await pb.collection('users').authWithOAuth2Code(
'google',
'CODE',
'VERIFIER',
'REDIRECT_URL',
// optional data that will be used for the new account on OAuth2 sign-up
'name': 'test',
// after the above you can also access the auth data from the authStore
console.log(pb.authStore.isValid);
console.log(pb.authStore.token);
console.log(pb.authStore.record.id);
// "logout" the last authenticated record
pb.authStore.clear();  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
final authData = await pb.collection('users').authWithOAuth2Code(
'google',
'CODE',
'VERIFIER',
'REDIRECT_URL',
// optional data that will be used for the new account on OAuth2 sign-up
createData: {
'name': 'test',
// after the above you can also access the auth data from the authStore
print(pb.authStore.isValid);
print(pb.authStore.token);
print(pb.authStore.record.id);
// "logout" the last authenticated record
pb.authStore.clear();   ###### API details
POST /api/collections/`collectionIdOrName`/auth-with-oauth2 Path parameters Param Type Description collectionIdOrName String ID or name of the auth collection. Body Parameters Param Type Description Required provider String The name of the OAuth2 client provider (e.g. &quot;google&quot;). Required code String The authorization code returned from the initial request. Required codeVerifier String The code verifier sent with the initial request as part of the code_challenge. Required redirectUrl String The redirect url sent with the initial request. Optional createData Object Optional data that will be used when creating the auth record on OAuth2 sign-up.
The created auth record must comply with the same requirements and validations in the
regular create action.
The data can only be in json, aka. multipart/form-data and
files upload currently are not supported during OAuth2 sign-ups.
Body parameters could be sent as JSON or
multipart/form-data. Query parameters Param Type Description expand String Auto expand record relations. Ex.:
`?expand=relField1,relField2.subRelField`
Supports up to 6-levels depth nested relations expansion.
The expanded relations will be appended to the record under the
`expand` property (e.g. `"expand": {"relField1": {...}, ...}`).
Only the relations to which the request user has permissions to view will be expanded. fields String Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,record.expand.relField.name
* targets all keys from the specific depth level.
In addition, the following field modifiers are also supported:
-:excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,record.description:excerpt(200,true)
Responses  {
"token": "eyJhbGciOiJIUzI1NiJ9.eyJpZCI6IjRxMXhsY2xtZmxva3UzMyIsInR5cGUiOiJhdXRoUmVjb3JkIiwiY29sbGVjdGlvbklkIjoiX3BiX3VzZXJzX2F1dGhfIiwiZXhwIjoyMjA4OTg1MjYxfQ.UwD8JvkbQtXpymT09d7J6fdA0aP9g4FJ1GPh_ggEkzc",
"record": {
"id": "8171022dc95a4ed",
"collectionId": "d2972397d45614e",
"collectionName": "users",
"created": "2022-06-24 06:24:18.434Z",
"updated": "2022-06-24 06:24:18.889Z",
"verified": true,
"emailVisibility": false,
"someCustomField": "example 123"
"meta": {
"id": "abc123",
"name": "John Doe",
"username": "john.doe",
"isNew": false,
"accessToken": "...",
"refreshToken": "...",
"expiry": "..."
}  {
"status": 400,
"message": "An error occurred while submitting the form.",
"data": {
"provider": {
"code": "validation_required",
"message": "Missing required value."
}      Auth with OTP    Authenticate a single auth record with an one-time password (OTP).
Note that when requesting an OTP we return an otpId even if a user with the provided email
doesn&#39;t exist as a very basic enumeration protection.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
// send OTP email to the provided auth record
// ... show a screen/popup to enter the password from the email ...
// authenticate with the requested OTP id and the email password
const authData = await pb.collection('users').authWithOTP(req.otpId, "YOUR_OTP");
// after the above you can also access the auth data from the authStore
console.log(pb.authStore.isValid);
console.log(pb.authStore.token);
console.log(pb.authStore.record.id);
// "logout"
pb.authStore.clear();  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
// send OTP email to the provided auth record
// ... show a screen/popup to enter the password from the email ...
// authenticate with the requested OTP id and the email password
final authData = await pb.collection('users').authWithOTP(req.otpId, "YOUR_OTP");
// after the above you can also access the auth data from the authStore
print(pb.authStore.isValid);
print(pb.authStore.token);
print(pb.authStore.record.id);
// "logout"
pb.authStore.clear();   ###### API details
OTP Auth  POST /api/collections/`collectionIdOrName`/request-otp Path parameters Param Type Description collectionIdOrName String ID or name of the auth collection. Body Parameters Param Type Description Required email String The auth record email address to send the OTP request (if exists). Responses  {
"otpId": "frJOeqFIiPIbaYT"
}  {
"status": 400,
"message": "An error occurred while validating the submitted data.",
"data": {
"email": {
"code": "validation_is_email",
"message": "Must be a valid email address."
}  {
"status": 429,
"message": "You've send too many OTP requests, please try again later.",
"data": {}
}   POST /api/collections/`collectionIdOrName`/auth-with-otp Path parameters Param Type Description collectionIdOrName String ID or name of the auth collection. Body Parameters Param Type Description Required otpId String The id of the OTP request. Required password String The one-time password. Query parameters Param Type Description expand String Auto expand record relations. Ex.:
`?expand=relField1,relField2.subRelField`
Supports up to 6-levels depth nested relations expansion.
The expanded relations will be appended to the record under the
`expand` property (e.g. `"expand": {"relField1": {...}, ...}`).
Only the relations to which the request user has permissions to view will be expanded. fields String Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,record.expand.relField.name
* targets all keys from the specific depth level.
In addition, the following field modifiers are also supported:
-:excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,record.description:excerpt(200,true)
Responses  {
"token": "eyJhbGciOiJIUzI1NiJ9.eyJpZCI6IjRxMXhsY2xtZmxva3UzMyIsInR5cGUiOiJhdXRoUmVjb3JkIiwiY29sbGVjdGlvbklkIjoiX3BiX3VzZXJzX2F1dGhfIiwiZXhwIjoyMjA4OTg1MjYxfQ.UwD8JvkbQtXpymT09d7J6fdA0aP9g4FJ1GPh_ggEkzc",
"record": {
"id": "8171022dc95a4ed",
"collectionId": "d2972397d45614e",
"collectionName": "users",
"created": "2022-06-24 06:24:18.434Z",
"updated": "2022-06-24 06:24:18.889Z",
"verified": false,
"emailVisibility": true,
"someCustomField": "example 123"
}  {
"status": 400,
"message": "Failed to authenticate.",
"data": {
"otpId": {
"code": "validation_required",
"message": "Missing required value."
}       Auth refresh    Returns a new auth response (token and user data) for already authenticated auth record.
stored data in pb.authStore is still valid and up-to-date.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
const authData = await pb.collection('users').authRefresh();
// after the above you can also access the refreshed auth data from the authStore
console.log(pb.authStore.isValid);
console.log(pb.authStore.token);
console.log(pb.authStore.record.id);  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
final authData = await pb.collection('users').authRefresh();
// after the above you can also access the refreshed auth data from the authStore
print(pb.authStore.isValid);
print(pb.authStore.token);
print(pb.authStore.record.id);   ###### API details
POST /api/collections/`collectionIdOrName`/auth-refresh Requires `Authorization:TOKEN` Path parameters Param Type Description collectionIdOrName String ID or name of the auth collection. Query parameters Param Type Description expand String Auto expand record relations. Ex.:
`?expand=relField1,relField2.subRelField`
Supports up to 6-levels depth nested relations expansion.
The expanded relations will be appended to the record under the
`expand` property (e.g. `"expand": {"relField1": {...}, ...}`).
Only the relations to which the request user has permissions to view will be expanded. fields String Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,record.expand.relField.name
* targets all keys from the specific depth level.
In addition, the following field modifiers are also supported:
-:excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,record.description:excerpt(200,true)
Responses  {
"token": "eyJhbGciOiJIUzI1NiJ9.eyJpZCI6IjRxMXhsY2xtZmxva3UzMyIsInR5cGUiOiJhdXRoUmVjb3JkIiwiY29sbGVjdGlvbklkIjoiX3BiX3VzZXJzX2F1dGhfIiwiZXhwIjoyMjA4OTg1MjYxfQ.UwD8JvkbQtXpymT09d7J6fdA0aP9g4FJ1GPh_ggEkzc",
"record": {
"id": "8171022dc95a4ed",
"collectionId": "d2972397d45614e",
"collectionName": "users",
"created": "2022-06-24 06:24:18.434Z",
"updated": "2022-06-24 06:24:18.889Z",
"verified": false,
"emailVisibility": true,
"someCustomField": "example 123"
}  {
"status": 401,
"message": "The request requires valid record authorization token to be set.",
"data": {}
}  {
"status": 403,
"message": "The authorized record model is not allowed to perform this action.",
"data": {}
}  {
"status": 404,
"message": "Missing auth record context.",
"data": {}
}      Verification    Sends auth record email verification request.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
// (optional) in your custom confirmation page:
await pb.collection('users').confirmVerification('VERIFICATION_TOKEN');  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
// (optional) in your custom confirmation page:
await pb.collection('users').confirmVerification('VERIFICATION_TOKEN');   ###### API details
Confirm verification  POST /api/collections/`collectionIdOrName`/request-verification Body Parameters Param Type Description Required email String The auth record email address to send the verification request (if exists). Responses  `null`  {
"status": 400,
"message": "An error occurred while validating the submitted data.",
"data": {
"email": {
"code": "validation_required",
"message": "Missing required value."
}   POST /api/collections/`collectionIdOrName`/confirm-verification Body Parameters Param Type Description Required token String The token from the verification request email. Responses  `null`  {
"status": 400,
"message": "An error occurred while validating the submitted data.",
"data": {
"token": {
"code": "validation_required",
"message": "Missing required value."
}       Password reset    Sends auth record password reset email request.
automatically invalidated.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
// (optional) in your custom confirmation page:
await pb.collection('users').confirmPasswordReset(
'RESET_TOKEN',
'NEW_PASSWORD',
'NEW_PASSWORD_CONFIRM',
);  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
// (optional) in your custom confirmation page:
await pb.collection('users').confirmPasswordReset(
'RESET_TOKEN',
'NEW_PASSWORD',
'NEW_PASSWORD_CONFIRM',
);   ###### API details
Confirm password reset  POST /api/collections/`collectionIdOrName`/request-password-reset Body Parameters Param Type Description Required email String The auth record email address to send the password reset request (if exists). Responses  `null`  {
"status": 400,
"message": "An error occurred while validating the submitted data.",
"data": {
"email": {
"code": "validation_required",
"message": "Missing required value."
}   POST /api/collections/`collectionIdOrName`/confirm-password-reset Body Parameters Param Type Description Required token String The token from the password reset request email. Required password String The new password to set. Required passwordConfirm String The new password confirmation. Responses  `null`  {
"status": 400,
"message": "An error occurred while validating the submitted data.",
"data": {
"token": {
"code": "validation_required",
"message": "Missing required value."
}       Email change    Sends auth record email change request.
automatically invalidated.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
// (optional) in your custom confirmation page:
await pb.collection('users').confirmEmailChange('EMAIL_CHANGE_TOKEN', 'YOUR_PASSWORD');  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
// (optional) in your custom confirmation page:
await pb.collection('users').confirmEmailChange('EMAIL_CHANGE_TOKEN', 'YOUR_PASSWORD');   ###### API details
Confirm email change  POST /api/collections/`collectionIdOrName`/request-email-change Requires Authorization:TOKEN
Body Parameters Param Type Description Required newEmail String The new email address to send the change email request. Responses  `null`  {
"status": 400,
"message": "An error occurred while validating the submitted data.",
"data": {
"newEmail": {
"code": "validation_required",
"message": "Missing required value."
}  {
"status": 401,
"message": "The request requires valid record authorization token to be set.",
"data": {}
}  {
"status": 403,
"message": "The authorized record model is not allowed to perform this action.",
"data": {}
}   POST /api/collections/`collectionIdOrName`/confirm-email-change Body Parameters Param Type Description Required token String The token from the change email request email. Required password String The account password to confirm the email change. Responses  `null`  {
"status": 400,
"message": "An error occurred while validating the submitted data.",
"data": {
"token": {
"code": "validation_required",
"message": "Missing required value."
}       Impersonate    Impersonate allows you to authenticate as a different user by generating a
nonrefreshable auth token.
Only superusers can perform this action.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
// authenticate as superuser
// impersonate
// (the custom token duration is optional and must be in seconds)
const impersonateClient = pb.collection("users").impersonate("USER_RECORD_ID", 3600)
// log the impersonate token and user data
console.log(impersonateClient.authStore.token);
console.log(impersonateClient.authStore.record);
// send requests as the impersonated user
impersonateClient.collection("example").getFullList();  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
// authenticate as superuser
// impersonate
// (the custom token duration is optional and must be in seconds)
final impersonateClient = pb.collection("users").impersonate("USER_RECORD_ID", 3600)
// log the impersonate token and user data
print(impersonateClient.authStore.token);
print(impersonateClient.authStore.record);
// send requests as the impersonated user
impersonateClient.collection("example").getFullList();   ###### API details
POST /api/collections/`collectionIdOrName`/impersonate/`id` Requires `Authorization:TOKEN` Path parameters Param Type Description collectionIdOrName String ID or name of the auth collection. id String ID of the auth record to impersonate. Body Parameters Param Type Description Optional duration Number Optional custom JWT duration for the `exp` claim (in seconds).
If not set or 0, it fallbacks to the default collection auth token duration option. Body parameters could be sent as JSON or
multipart/form-data. Query parameters Param Type Description expand String Auto expand record relations. Ex.:
`?expand=relField1,relField2.subRelField`
Supports up to 6-levels depth nested relations expansion.
The expanded relations will be appended to the record under the
`expand` property (e.g. `"expand": {"relField1": {...}, ...}`).
Only the relations to which the request user has permissions to view will be expanded. fields String Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,record.expand.relField.name
* targets all keys from the specific depth level.
In addition, the following field modifiers are also supported:
-:excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,record.description:excerpt(200,true)
Responses  {
"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjb2xsZWN0aW9uSWQiOiJfcGJjX2MwcHdrZXNjcXMiLCJleHAiOjE3MzAzNjgxMTUsImlkIjoicXkwMmMxdDBueDBvanFuIiwicmVmcmVzaGFibGUiOmZhbHNlLCJ0eXBlIjoiYXV0aCJ9.1JOaE54TyPdDLf0mb0T6roIYeh8Y1HfJvDlYZADMN4U",
"record": {
"id": "8171022dc95a4ed",
"collectionId": "d2972397d45614e",
"collectionName": "users",
"created": "2022-06-24 06:24:18.434Z",
"updated": "2022-06-24 06:24:18.889Z",
"verified": false,
"emailVisibility": true,
"someCustomField": "example 123"
}  {
"status": 400,
"message": "The request requires valid record authorization token to be set.",
"data": {
"duration": {
"code": "validation_min_greater_equal_than_required",
"message": "Must be no less than 0."
}  {
"status": 401,
"message": "An error occurred while validating the submitted data.",
"data": {}
}  {
"status": 403,
"message": "The authorized record model is not allowed to perform this action.",
"data": {}
}  {
"status": 404,
"message": "The requested resource wasn't found.",
"data": {}

## 18.Web APIs reference - API Realtime
Required clientId(String):ID of the SSE client connection.
Optional subscriptions(Array<String>):The new client subscriptions to set in the format:
COLLECTION_ID_OR_NAME or
COLLECTION_ID_OR_NAME/RECORD_ID. You can also attach optional query and header parameters as serialized json to a
single topic using the options
query parameter, e.g.:
COLLECTION_ID_OR_NAME/RECORD_ID?options={"query": {"abc": "123"}, "headers": {"x-token": "..."}} Leave empty to unsubscribe from everything.
# Web APIs reference - API Realtime
- **Required clientId** (String): ID of the SSE client connection.
- **Optional subscriptions** (Array<String>): The new client subscriptions to set in the format:
COLLECTION_ID_OR_NAME or
COLLECTION_ID_OR_NAME/RECORD_ID. You can also attach optional query and header parameters as serialized json to a
single topic using the options
query parameter, e.g.:
COLLECTION_ID_OR_NAME/RECORD_ID?options={"query": {"abc": "123"}, "headers": {"x-token": "..."}} Leave empty to unsubscribe from everything.
API Realtime The Realtime API is implemented via Server-Sent Events (SSE). Generally, it consists of 2 operations:
-establish SSE connection
-submit client subscriptions
SSE events are sent for create, update
and delete record operations.
You could subscribe to a single record or to an entire collection.
When you subscribe to a single record, the collection&#39;s
ViewRule will be used to determine whether the subscriber has access to receive the
event message.
When you subscribe to an entire collection, the collection&#39;s
ListRule will be used to determine whether the subscriber has access to receive the
event message.
Connect    GET /api/realtime Establishes a new SSE connection and immediately sends a PB_CONNECT SSE event with the
created client ID.
NB! The user/superuser authorization happens during the first
Set subscriptions
call.
If the connected client doesn&#39;t receive any new messages for 5 minutes, the server will send a
disconnect signal (this is to prevent forgotten/leaked connections). The connection will be
automatically reestablished if the client is still active (e.g. the browser tab is still open).
If Authorization header is set, will authorize the client SSE connection with the
associated user or superuser.
Body Parameters Param Type Description Required clientId String ID of the SSE client connection. Optional subscriptions Array&lt;String> The new client subscriptions to set in the format:
COLLECTION_ID_OR_NAME or
COLLECTION_ID_OR_NAME/RECORD_ID.
You can also attach optional query and header parameters as serialized json to a
single topic using the options
query parameter, e.g.:
COLLECTION_ID_OR_NAME/RECORD_ID?options={"query": {"abc": "123"}, "headers": {"x-token": "..."}}
Leave empty to unsubscribe from everything.
Body parameters could be sent as JSON or
multipart/form-data. Responses  `null`  {
"status": 400,
"message": "Something went wrong while processing your request.",
"data": {
"clientId": {
"code": "validation_required",
"message": "Missing required value."
}  {
"status": 403,
"data": {}
}  {
"status": 404,
"message": "Missing or invalid client id.",
"data": {}
}   All of this is seamlessly handled by the SDKs using just the subscribe and
unsubscribe methods:
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
// (Optionally) authenticate
// Subscribe to changes in any record in the collection
pb.collection('example').subscribe('*', function (e) {
console.log(e.action);
console.log(e.record);
}, { /* other options like expand, custom headers, etc. */ });
// Subscribe to changes only in the specified record
pb.collection('example').subscribe('RECORD_ID', function (e) {
console.log(e.action);
console.log(e.record);
}, { /* other options like expand, custom headers, etc. */ });
// Unsubscribe
pb.collection('example').unsubscribe('RECORD_ID'); // remove all 'RECORD_ID' subscriptions
pb.collection('example').unsubscribe('*'); // remove all '*' topic subscriptions
pb.collection('example').unsubscribe(); // remove all subscriptions in the collection  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
// (Optionally) authenticate
// Subscribe to changes in any record in the collection
pb.collection('example').subscribe('*', (e) {
print(e.action);
print(e.record);
}, /* other options like expand, custom headers, etc. */);
// Subscribe to changes only in the specified record
pb.collection('example').subscribe('RECORD_ID', (e) {
print(e.action);
print(e.record);
}, /* other options like expand, custom headers, etc. */);
// Unsubscribe
pb.collection('example').unsubscribe('RECORD_ID'); // remove all 'RECORD_ID' subscriptions
pb.collection('example').unsubscribe('*'); // remove all '*' topic subscriptions
pb.collection('example').unsubscribe(); // remove all subscriptions in the collection

## 19.Web APIs reference - API Files
collectionIdOrName(String):ID or name of the collection whose record model contains the file resource.
recordId(String):ID of the record model that contains the file resource.
filename(String):Name of the file resource.
thumb(String):Get the thumb of the requested file.
The following thumb formats are currently supported: WxH
(e.g. 100x300) - crop to WxH viewbox (from center) WxHt
(e.g. 100x300t) - crop to WxH viewbox (from top) WxHb
(e.g. 100x300b) - crop to WxH viewbox (from bottom) WxHf
(e.g. 100x300f) - fit inside a WxH viewbox (without cropping) 0xH
(e.g. 0x300) - resize to H height preserving the aspect ratio Wx0
(e.g. 100x0) - resize to W width preserving the aspect ratio
If the thumb size is not defined in the file schema field options or the file resource is not
an image (jpg, png, gif, webp), then the original file resource is returned unmodified.
token(String):Optional file token for granting access to
protected file(s).
For an example, you can check
"Files upload and handling".
download(Boolean):If it is set to a truthy value (1, t, true) the file will be
served with Content-Disposition: attachment header instructing the browser to
ignore the file preview for pdf, images, videos, etc. and to directly download the file.
# Web APIs reference - API Files
- **collectionIdOrName** (String): ID or name of the collection whose record model contains the file resource.
- **recordId** (String): ID of the record model that contains the file resource.
- **filename** (String): Name of the file resource.
- **thumb** (String): Get the thumb of the requested file.
The following thumb formats are currently supported: WxH
(e.g. 100x300) - crop to WxH viewbox (from center) WxHt
(e.g. 100x300t) - crop to WxH viewbox (from top) WxHb
(e.g. 100x300b) - crop to WxH viewbox (from bottom) WxHf
(e.g. 100x300f) - fit inside a WxH viewbox (without cropping) 0xH
(e.g. 0x300) - resize to H height preserving the aspect ratio Wx0
(e.g. 100x0) - resize to W width preserving the aspect ratio
If the thumb size is not defined in the file schema field options or the file resource is not
an image (jpg, png, gif, webp), then the original file resource is returned unmodified.
- **token** (String): Optional file token for granting access to
protected file(s).
For an example, you can check
"Files upload and handling".
served with Content-Disposition: attachment header instructing the browser to
API Files Files are uploaded, updated or deleted via the
Records API.
manipulations, like generating thumbs).
GET  /api/files/`collectionIdOrName`/`recordId`/`filename` Path parameters Param Type Description collectionIdOrName String ID or name of the collection whose record model contains the file resource. recordId String ID of the record model that contains the file resource. filename String Name of the file resource. Query parameters Param Type Description thumb String Get the thumb of the requested file.
The following thumb formats are currently supported:
-WxH
(e.g. 100x300) - crop to WxH viewbox (from center)
-WxHt
(e.g. 100x300t) - crop to WxH viewbox (from top)
-WxHb
(e.g. 100x300b) - crop to WxH viewbox (from bottom)
-WxHf
(e.g. 100x300f) - fit inside a WxH viewbox (without cropping)
-0xH
(e.g. 0x300) - resize to H height preserving the aspect ratio
-Wx0
(e.g. 100x0) - resize to W width preserving the aspect ratio
If the thumb size is not defined in the file schema field options or the file resource is not
an image (jpg, png, gif, webp), then the original file resource is returned unmodified. token String Optional file token for granting access to
protected file(s).
For an example, you can check
served with `Content-Disposition: attachment` header instructing the browser to
"status": 400,
"message": "Filesystem initialization failure.",
"data": {}
}  {
"status": 404,
"message": "The requested resource wasn't found.",
"data": {}
}      Generate protected file token    Generates a short-lived file token for accessing
protected file(s).
The client must be superuser or auth record authenticated (aka. have regular authorization token
sent with the request).
###### API details
POST /api/files/token Requires `Authorization:TOKEN` Responses  {
"token": "..."
}  {
"status": 400,
"message": "Failed to generate file token.",
"data": {}

## 20.Web APIs reference - API Crons
fields(String):Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
jobId(String):The identifier of the cron job to run.
# Web APIs reference - API Crons
- **fields** (String): Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
- **jobId** (String): The identifier of the cron job to run.
API Crons    List cron jobs    Returns list with all registered app level cron jobs.
Only superusers can perform this action.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
const jobs = await pb.crons.getFullList();  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
final jobs = await pb.crons.getFullList();   ###### API details
GET /api/crons Query parameters Param Type Description fields String Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name
* targets all keys from the specific depth level.
In addition, the following field modifiers are also supported:
-:excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
Responses  [
"id": "__pbDBOptimize__",
"expression": "0 0 * * *"
"id": "__pbMFACleanup__",
"expression": "0 * * * *"
"id": "__pbOTPCleanup__",
"expression": "0 * * * *"
"id": "__pbLogsCleanup__",
"expression": "0 */6 * * *"
]  {
"status": 400,
"message": "Failed to load backups filesystem.",
"data": {}
}  {
"status": 401,
"message": "The request requires valid record authorization token.",
"data": {}
}  {
"status": 403,
"message": "Only superusers can perform this action.",
"data": {}
}      Run cron job    Triggers a single cron job by its id.
Only superusers can perform this action.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
await pb.crons.run('__pbLogsCleanup__');  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
await pb.crons.run('__pbLogsCleanup__');   ###### API details
POST /api/crons/`jobId` Requires `Authorization:TOKEN` Path parameters Param Type Description jobId String The identifier of the cron job to run. Responses  `null`  {
"status": 401,
"message": "The request requires valid record authorization token.",
"data": {}
}  {
"status": 403,
"message": "The authorized record is not allowed to perform this action.",
"data": {}
}  {
"status": 404,
"message": "Missing or invalid cron job.",
"data": {}

## 21.Web APIs reference - API Settings
fields(String):Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
├─ Required appName(String):The app name.
├─ Required appUrl(String):The app public absolute url.
├─ Optional hideControls(Boolean):Hides the collection create and update controls from the Dashboard.
Useful to prevent making accidental schema changes when in production environment.
├─ Required senderName(String):Transactional mails sender name.
├─ Required senderAddress(String):Transactional mails sender address.
└─ Optional maxDays(Number):Max retention period. Set to 0 for no logs.
└─ Optional minLevel(Number):Specifies the minimum log persistent level.
The default log levels are:
-4: DEBUG 0: INFO 4: WARN 8: ERROR
└─ Optional logIP(Boolean):If enabled includes the client IP in the activity request logs.
└─ Optional logAuthId(Boolean):If enabled includes the authenticated record id in the activity request logs.
├─ Optional cron(String):Cron expression to schedule auto backups, e.g. 0 0 * * *.
├─ Optional cronMaxKeep(Number):The max number of cron generated backups to keep before removing older entries.
└─ Optional s3(Object):S3 configuration (the same fields as for the S3 file storage settings).
├─ Optional enabled(Boolean):Enable the use of the SMTP mail server for sending emails.
├─ Required host*(String):Mail server host (required if SMTP is enabled).
├─ Required port*(Number):Mail server port (required if SMTP is enabled).
├─ Optional username(String):Mail server username.
├─ Optional password(String):Mail server password.
├─ Optional tls(Boolean):Whether to enforce TLS connection encryption.
When false StartTLS command is send, leaving the server to decide whether
to upgrade the connection or not).
├─ Optional authMethod(String):The SMTP AUTH method to use - PLAIN or LOGIN (used mainly by Microsoft).
Default to PLAIN if empty.
└─ Optional localName(String):Optional domain name or (IP address) to use for the initial EHLO/HELO exchange.
If not explicitly set, localhost will be used.
Note that some SMTP providers, such as Gmail SMTP-relay, requires a proper domain name and
and will reject attempts to use localhost.
├─ Optional enabled(Boolean):Enable the use of a S3 compatible storage.
├─ Required bucket*(String):S3 storage bucket (required if enabled).
├─ Required region*(String):S3 storage region (required if enabled).
├─ Required endpoint*(String):S3 storage public endpoint (required if enabled).
├─ Required accessKey*(String):S3 storage access key (required if enabled).
├─ Required secret*(String):S3 storage secret (required if enabled).
└─ Optional forcePathStyle(Boolean):Forces the S3 request to use path-style addressing, e.g.
"https://s3.amazonaws.com/BUCKET/KEY" instead of the default
"https://BUCKET.s3.amazonaws.com/KEY".
├─ Optional enabled(Boolean):Enable the batch Web APIs.
├─ Required maxRequests(Number):The maximum allowed batch request to execute.
├─ Required timeout(Number):The max duration in seconds to wait before cancelling the batch transaction.
└─ Optional maxBodySize(Number):The maximum allowed batch request body size in bytes.
If not set, fallbacks to max ~128MB.
├─ Optional enabled(Boolean):Enable the builtin rate limiter.
└─ Optional rules(Array<RateLimitRule>):List of rate limit rules. Each rule have:
label - the identifier of the rule.
It could be a tag, complete path or path prerefix (when ends with `/`). maxRequests - the max allowed number of requests per duration. duration - specifies the interval (in seconds) per which to reset the
counted/accumulated rate limiter tokens..
├─ Optional headers(Array<String>):List of explicit trusted header(s) to check.
└─ Optional useLeftmostIP(Boolean):Specifies to use the left-mostish IP from the trusted headers.
fields(String):Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
Required filesystem(String):The storage filesystem to test (storage or backups).
Optional collection(String):The name or id of the auth collection. Fallbacks to _superusers if not set.
Required email(String):The receiver of the test email.
Required template(String):The test email template to send: verification,
password-reset or
email-change.
Required clientId(String):The identifier of your app (aka. Service ID).
Required teamId(String):10-character string associated with your developer account (usually could be found next to
your name in the Apple Developer site).
Required keyId(String):10-character key identifier generated for the "Sign in with Apple" private key associated
with your developer account.
Required privateKey(String):PrivateKey is the private key associated to your app.
Required duration(Number):Duration specifies how long the generated JWT token should be considered valid.
The specified value must be in seconds and max 15777000 (~6months).
# Web APIs reference - API Settings
- **fields** (String): Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
- **├─ Required appName** (String): The app name.
- **├─ Required appUrl** (String): The app public absolute url.
- **├─ Optional hideControls** (Boolean): Hides the collection create and update controls from the Dashboard.
Useful to prevent making accidental schema changes when in production environment.
- **├─ Required senderName** (String): Transactional mails sender name.
- **├─ Required senderAddress** (String): Transactional mails sender address.
- **└─ Optional maxDays** (Number): Max retention period. Set to 0 for no logs.
- **└─ Optional minLevel** (Number): Specifies the minimum log persistent level.
The default log levels are:
-4: DEBUG 0: INFO 4: WARN 8: ERROR
- **└─ Optional logIP** (Boolean): If enabled includes the client IP in the activity request logs.
- **└─ Optional logAuthId** (Boolean): If enabled includes the authenticated record id in the activity request logs.
- **├─ Optional cron** (String): Cron expression to schedule auto backups, e.g. 0 0 * * *.
- **├─ Optional cronMaxKeep** (Number): The max number of cron generated backups to keep before removing older entries.
- **└─ Optional s3** (Object): S3 configuration (the same fields as for the S3 file storage settings).
- **├─ Optional enabled** (Boolean): Enable the use of the SMTP mail server for sending emails.
- **├─ Required host** (String) (required): Mail server host (required if SMTP is enabled).
- **├─ Required port** (Number) (required): Mail server port (required if SMTP is enabled).
- **├─ Optional username** (String): Mail server username.
- **├─ Optional password** (String): Mail server password.
- **├─ Optional tls** (Boolean): Whether to enforce TLS connection encryption.
When false StartTLS command is send, leaving the server to decide whether
to upgrade the connection or not).
- **├─ Optional authMethod** (String): The SMTP AUTH method to use - PLAIN or LOGIN (used mainly by Microsoft).
Default to PLAIN if empty.
- **└─ Optional localName** (String): Optional domain name or (IP address) to use for the initial EHLO/HELO exchange.
If not explicitly set, localhost will be used.
Note that some SMTP providers, such as Gmail SMTP-relay, requires a proper domain name and
and will reject attempts to use localhost.
- **├─ Optional enabled** (Boolean): Enable the use of a S3 compatible storage.
- **├─ Required bucket** (String) (required): S3 storage bucket (required if enabled).
- **├─ Required region** (String) (required): S3 storage region (required if enabled).
- **├─ Required endpoint** (String) (required): S3 storage public endpoint (required if enabled).
- **├─ Required accessKey** (String) (required): S3 storage access key (required if enabled).
- **├─ Required secret** (String) (required): S3 storage secret (required if enabled).
- **└─ Optional forcePathStyle** (Boolean): Forces the S3 request to use path-style addressing, e.g.
"https://s3.amazonaws.com/BUCKET/KEY" instead of the default
"https://BUCKET.s3.amazonaws.com/KEY".
- **├─ Optional enabled** (Boolean): Enable the batch Web APIs.
- **├─ Required maxRequests** (Number): The maximum allowed batch request to execute.
- **├─ Required timeout** (Number): The max duration in seconds to wait before cancelling the batch transaction.
- **└─ Optional maxBodySize** (Number): The maximum allowed batch request body size in bytes.
If not set, fallbacks to max ~128MB.
- **├─ Optional enabled** (Boolean): Enable the builtin rate limiter.
- **└─ Optional rules** (Array<RateLimitRule>): List of rate limit rules. Each rule have:
label - the identifier of the rule.
It could be a tag, complete path or path prerefix (when ends with `/`). maxRequests - the max allowed number of requests per duration. duration - specifies the interval (in seconds) per which to reset the
counted/accumulated rate limiter tokens..
- **├─ Optional headers** (Array<String>): List of explicit trusted header(s) to check.
- **└─ Optional useLeftmostIP** (Boolean): Specifies to use the left-mostish IP from the trusted headers.
- **fields** (String): Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
- **Required filesystem** (String): The storage filesystem to test (storage or backups).
- **Optional collection** (String): The name or id of the auth collection. Fallbacks to _superusers if not set.
- **Required email** (String): The receiver of the test email.
- **Required template** (String): The test email template to send: verification,
password-reset or
email-change.
- **Required clientId** (String): The identifier of your app (aka. Service ID).
your name in the Apple Developer site).
- **Required keyId** (String): 10-character key identifier generated for the "Sign in with Apple" private key associated
with your developer account.
- **Required privateKey** (String): PrivateKey is the private key associated to your app.
- **Required duration** (Number): Duration specifies how long the generated JWT token should be considered valid.
The specified value must be in seconds and max 15777000 (~6months).
API Settings    List settings    Returns a list with all available application settings.
Secret/password fields are automatically redacted with ****** characters.
Only superusers can perform this action.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
const settings = await pb.settings.getAll();  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
final settings = await pb.settings.getAll();   ###### API details
GET /api/settings Requires `Authorization:TOKEN` Query parameters Param Type Description fields String Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name
* targets all keys from the specific depth level.
In addition, the following field modifiers are also supported:
-:excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
Responses  {
"smtp": {
"enabled": false,
"port": 587,
"username": "",
"authMethod": "",
"tls": true,
"localName": ""
"backups": {
"cron": "0 0 * * *",
"cronMaxKeep": 3,
"s3": {
"enabled": false,
"bucket": "",
"region": "",
"endpoint": "",
"accessKey": "",
"forcePathStyle": false
"s3": {
"enabled": false,
"bucket": "",
"region": "",
"endpoint": "",
"accessKey": "",
"forcePathStyle": false
"meta": {
"appName": "Acme",
"senderName": "Support",
"hideControls": false
"rateLimits": {
"rules": [
"label": "*:auth",
"audience": "",
"duration": 3,
"maxRequests": 2
"label": "*:create",
"audience": "",
"duration": 5,
"maxRequests": 20
"label": "/api/batch",
"audience": "",
"duration": 1,
"maxRequests": 3
"label": "/api/",
"audience": "",
"duration": 10,
"maxRequests": 300
"enabled": false
"trustedProxy": {
"headers": [],
"useLeftmostIP": false
"batch": {
"enabled": true,
"maxRequests": 50,
"timeout": 3,
"maxBodySize": 0
"logs": {
"maxDays": 7,
"minLevel": 0,
"logIP": true,
"logAuthId": false
}  {
"status": 401,
"message": "The request requires valid record authorization token.",
"data": {}
}  {
"status": 403,
"message": "The authorized record is not allowed to perform this action.",
"data": {}
}      Update settings    Bulk updates application settings and returns the updated settings list.
Only superusers can perform this action.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
const settings = await pb.settings.update({
meta: {
appName: 'YOUR_APP',
appUrl: 'http://127.0.0.1:8090',
});  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
final settings = await pb.settings.update(body: {
'meta': {
'appName': 'YOUR_APP',
'appUrl': 'http://127.0.0.1:8090',
});   ###### API details
PATCH /api/settings Requires `Authorization:TOKEN` Body Parameters Param Type Description  meta  Application meta data (name, url, support email, etc.). ├─ Required appName String The app name. ├─ Required appUrl String The app public absolute url. ├─ Optional hideControls Boolean Hides the collection create and update controls from the Dashboard.
Useful to prevent making accidental schema changes when in production environment. ├─ Required senderName String Transactional mails sender name. ├─ Required senderAddress String Transactional mails sender address.  logs  App logger settings. └─ Optional maxDays Number Max retention period. Set to 0 for no logs. └─ Optional minLevel Number Specifies the minimum log persistent level.
The default log levels are:
--4: DEBUG
-0: INFO
-4: WARN
-8: ERROR
└─ Optional logIP Boolean If enabled includes the client IP in the activity request logs. └─ Optional logAuthId Boolean If enabled includes the authenticated record id in the activity request logs.  backups  App data backups settings. ├─ Optional cron String Cron expression to schedule auto backups, e.g. `0 0 * * *`. ├─ Optional cronMaxKeep Number The max number of cron generated backups to keep before removing older entries. └─ Optional s3 Object S3 configuration (the same fields as for the S3 file storage settings).  smtp  SMTP mail server settings. ├─ Optional enabled Boolean Enable the use of the SMTP mail server for sending emails. ├─ Required host String Mail server host (required if SMTP is enabled). ├─ Required port Number Mail server port (required if SMTP is enabled). ├─ Optional username String Mail server username. ├─ Optional password String Mail server password. ├─ Optional tls Boolean Whether to enforce TLS connection encryption.
When false StartTLS command is send, leaving the server to decide whether
to upgrade the connection or not). ├─ Optional authMethod String The SMTP AUTH method to use - PLAIN or LOGIN (used mainly by Microsoft).
Default to PLAIN if empty. └─ Optional localName String Optional domain name or (IP address) to use for the initial EHLO/HELO exchange.
If not explicitly set, `localhost` will be used.
Note that some SMTP providers, such as Gmail SMTP-relay, requires a proper domain name and
and will reject attempts to use localhost.  s3  S3 compatible file storage settings. ├─ Optional enabled Boolean Enable the use of a S3 compatible storage. ├─ Required bucket String S3 storage bucket (required if enabled). ├─ Required region String S3 storage region (required if enabled). ├─ Required endpoint String S3 storage public endpoint (required if enabled). ├─ Required accessKey String S3 storage access key (required if enabled). ├─ Required secret String S3 storage secret (required if enabled). └─ Optional forcePathStyle Boolean Forces the S3 request to use path-style addressing, e.g.
&quot;https://s3.amazonaws.com/BUCKET/KEY&quot; instead of the default
&quot;https://BUCKET.s3.amazonaws.com/KEY&quot;.  batch  Batch logs settings. ├─ Optional enabled Boolean Enable the batch Web APIs. ├─ Required maxRequests Number The maximum allowed batch request to execute. ├─ Required timeout Number The max duration in seconds to wait before cancelling the batch transaction. └─ Optional maxBodySize Number The maximum allowed batch request body size in bytes.
If not set, fallbacks to max ~128MB.  rateLimits  Rate limiter settings. ├─ Optional enabled Boolean Enable the builtin rate limiter. └─ Optional rules Array&lt;RateLimitRule> List of rate limit rules. Each rule have:
-label - the identifier of the rule.
It could be a tag, complete path or path prerefix (when ends with `/`).
-maxRequests - the max allowed number of requests per duration.
-duration - specifies the interval (in seconds) per which to reset the
counted/accumulated rate limiter tokens..
trustedProxy  Trusted proxy headers settings. ├─ Optional headers Array&lt;String> List of explicit trusted header(s) to check. └─ Optional useLeftmostIP Boolean Specifies to use the left-mostish IP from the trusted headers. Body parameters could be sent as JSON or
multipart/form-data. Query parameters Param Type Description fields String Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name
* targets all keys from the specific depth level.
In addition, the following field modifiers are also supported:
-:excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
Responses  {
"smtp": {
"enabled": false,
"port": 587,
"username": "",
"authMethod": "",
"tls": true,
"localName": ""
"backups": {
"cron": "0 0 * * *",
"cronMaxKeep": 3,
"s3": {
"enabled": false,
"bucket": "",
"region": "",
"endpoint": "",
"accessKey": "",
"forcePathStyle": false
"s3": {
"enabled": false,
"bucket": "",
"region": "",
"endpoint": "",
"accessKey": "",
"forcePathStyle": false
"meta": {
"appName": "Acme",
"senderName": "Support",
"hideControls": false
"rateLimits": {
"rules": [
"label": "*:auth",
"audience": "",
"duration": 3,
"maxRequests": 2
"label": "*:create",
"audience": "",
"duration": 5,
"maxRequests": 20
"label": "/api/batch",
"audience": "",
"duration": 1,
"maxRequests": 3
"label": "/api/",
"audience": "",
"duration": 10,
"maxRequests": 300
"enabled": false
"trustedProxy": {
"headers": [],
"useLeftmostIP": false
"batch": {
"enabled": true,
"maxRequests": 50,
"timeout": 3,
"maxBodySize": 0
"logs": {
"maxDays": 7,
"minLevel": 0,
"logIP": true,
"logAuthId": false
}  {
"status": 400,
"message": "An error occurred while submitting the form.",
"data": {
"meta": {
"appName": {
"code": "validation_required",
"message": "Missing required value."
}  {
"status": 401,
"message": "The request requires valid record authorization token.",
"data": {}
}  {
"status": 403,
"message": "The authorized record is not allowed to perform this action.",
"data": {}
}      Test S3 storage connection    Performs S3 storage connection test.
Only superusers can perform this action.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
await pb.settings.testS3("backups");  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
await pb.settings.testS3("backups");   ###### API details
POST /api/settings/test/s3 Requires `Authorization:TOKEN` Body Parameters Param Type Description Required filesystem String The storage filesystem to test (`storage` or `backups`). Body parameters could be sent as JSON or
multipart/form-data. Responses  `null`  {
"status": 400,
"data": {}
}  {
"status": 401,
"message": "The request requires valid record authorization token.",
"data": {}
}      Send test email    Sends a test user email.
Only superusers can perform this action.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
final pb = PocketBase('http://127.0.0.1:8090');
POST /api/settings/test/email Requires `Authorization:TOKEN` Body Parameters Param Type Description Optional collection String The name or id of the auth collection. Fallbacks to _superusers if not set. Required email String The receiver of the test email. Required template String The test email template to send:  `verification`,
`password-reset` or
`email-change`. Body parameters could be sent as JSON or
multipart/form-data. Responses  `null`  {
"status": 400,
"message": "Failed to send the test email.",
"data": {
"email": {
"code": "validation_required",
"message": "Missing required value."
}  {
"status": 401,
"message": "The request requires valid record authorization token.",
"data": {}
}      Generate Apple client secret    Generates a new Apple OAuth2 client secret key.
Only superusers can perform this action.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
await pb.settings.generateAppleClientSecret(clientId, teamId, keyId, privateKey, duration)  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
await pb.settings.generateAppleClientSecret(clientId, teamId, keyId, privateKey, duration)   ###### API details
your name in the Apple Developer site). Required keyId String 10-character key identifier generated for the &quot;Sign in with Apple&quot; private key associated
with your developer account. Required privateKey String PrivateKey is the private key associated to your app. Required duration Number Duration specifies how long the generated JWT token should be considered valid.
The specified value must be in seconds and max 15777000 (~6months). Body parameters could be sent as JSON or
multipart/form-data. Responses  {
"secret": "..."
}  {
"status": 400,
"data": {}
}  {
"status": 401,
"message": "The request requires valid record authorization token.",
"data": {}

## 22.Web APIs reference - API Logs
GET /api/collections/_pbc_2287844090/records?page=1&amp;perPage=1&amp;filter=&amp;fields=id"
page(Number):The page (aka. offset) of the paginated list (default to 1).
perPage(Number):The max returned logs per page (default to 30).
sort(String):Specify the ORDER BY fields. Add - / + (default) in front of the attribute for DESC /
ASC order, e.g.: // DESC by the insertion rowid and ASC by level
?sort=-rowid,level Supported log sort fields: @random, rowid, id, created,
updated, level, message and any
data.* attribute.
filter(String):Filter expression to filter/search the returned logs list, e.g.: ?filter=(data.url~'test.com' && level>0) Supported log filter fields: id, created, updated,
level, message and any data.* attribute. The syntax basically follows the format
OPERAND OPERATOR OPERAND, where: OPERAND - could be any field literal, string (single or double quoted),
number, null, true, false OPERATOR - is one of:
= Equal != NOT equal > Greater than >= Greater than or equal < Less than <= Less than or equal ~ Like/Contains (if not specified auto wraps the right string OPERAND in a "%" for wildcard
match) !~ NOT Like/Contains (if not specified auto wraps the right string OPERAND in a "%" for
wildcard match) ?= Any/At least one of Equal ?!= Any/At least one of NOT equal ?> Any/At least one of Greater than ?>= Any/At least one of Greater than or equal ?< Any/At least one of Less than ?<= Any/At least one of Less than or equal ?~ Any/At least one of Like/Contains (if not specified auto wraps the right string OPERAND in a "%" for wildcard
match) ?!~ Any/At least one of NOT Like/Contains (if not specified auto wraps the right string OPERAND in a "%" for
wildcard match) To group and combine several expressions you can use parenthesis
(...), && (AND) and || (OR) tokens. Single line comments are also supported: // Example comment.
fields(String):Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
id(String):ID of the log to view.
fields(String):Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
filter(String):Filter expression to filter/search the logs, e.g.: ?filter=(data.url~'test.com' && level>0) Supported log filter fields: rowid, id, created,
updated, level, message and any
data.* attribute. The syntax basically follows the format
OPERAND OPERATOR OPERAND, where: OPERAND - could be any field literal, string (single or double quoted),
number, null, true, false OPERATOR - is one of:
= Equal != NOT equal > Greater than >= Greater than or equal < Less than <= Less than or equal ~ Like/Contains (if not specified auto wraps the right string OPERAND in a "%" for wildcard
match) !~ NOT Like/Contains (if not specified auto wraps the right string OPERAND in a "%" for
wildcard match) ?= Any/At least one of Equal ?!= Any/At least one of NOT equal ?> Any/At least one of Greater than ?>= Any/At least one of Greater than or equal ?< Any/At least one of Less than ?<= Any/At least one of Less than or equal ?~ Any/At least one of Like/Contains (if not specified auto wraps the right string OPERAND in a "%" for wildcard
match) ?!~ Any/At least one of NOT Like/Contains (if not specified auto wraps the right string OPERAND in a "%" for
wildcard match) To group and combine several expressions you can use parenthesis
(...), && (AND) and || (OR) tokens. Single line comments are also supported: // Example comment.
fields(String):Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
# Web APIs reference - API Logs
`GET /api/collections/_pbc_2287844090/records?page=1&amp;perPage=1&amp;filter=&amp;fields=id"`
- **page** (Number): The page (aka. offset) of the paginated list (default to 1).
- **perPage** (Number): The max returned logs per page (default to 30).
- **sort** (String): Specify the ORDER BY fields. Add - / + (default) in front of the attribute for DESC /
ASC order, e.g.: // DESC by the insertion rowid and ASC by level
?sort=-rowid,level Supported log sort fields: @random, rowid, id, created,
updated, level, message and any
data.* attribute.
- **filter** (String): Filter expression to filter/search the returned logs list, e.g.: ?filter=(data.url~'test.com' && level>0) Supported log filter fields: id, created, updated,
level, message and any data.* attribute. The syntax basically follows the format
OPERAND OPERATOR OPERAND, where: OPERAND - could be any field literal, string (single or double quoted),
number, null, true, false OPERATOR - is one of:
= Equal != NOT equal > Greater than >= Greater than or equal < Less than <= Less than or equal ~ Like/Contains (if not specified auto wraps the right string OPERAND in a "%" for wildcard
match) !~ NOT Like/Contains (if not specified auto wraps the right string OPERAND in a "%" for
wildcard match) ?= Any/At least one of Equal ?!= Any/At least one of NOT equal ?> Any/At least one of Greater than ?>= Any/At least one of Greater than or equal ?< Any/At least one of Less than ?<= Any/At least one of Less than or equal ?~ Any/At least one of Like/Contains (if not specified auto wraps the right string OPERAND in a "%" for wildcard
match) ?!~ Any/At least one of NOT Like/Contains (if not specified auto wraps the right string OPERAND in a "%" for
wildcard match) To group and combine several expressions you can use parenthesis
(...), && (AND) and || (OR) tokens. Single line comments are also supported: // Example comment.
- **fields** (String): Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
- **id** (String): ID of the log to view.
- **fields** (String): Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
- **filter** (String): Filter expression to filter/search the logs, e.g.: ?filter=(data.url~'test.com' && level>0) Supported log filter fields: rowid, id, created,
updated, level, message and any
data.* attribute. The syntax basically follows the format
OPERAND OPERATOR OPERAND, where: OPERAND - could be any field literal, string (single or double quoted),
number, null, true, false OPERATOR - is one of:
= Equal != NOT equal > Greater than >= Greater than or equal < Less than <= Less than or equal ~ Like/Contains (if not specified auto wraps the right string OPERAND in a "%" for wildcard
match) !~ NOT Like/Contains (if not specified auto wraps the right string OPERAND in a "%" for
wildcard match) ?= Any/At least one of Equal ?!= Any/At least one of NOT equal ?> Any/At least one of Greater than ?>= Any/At least one of Greater than or equal ?< Any/At least one of Less than ?<= Any/At least one of Less than or equal ?~ Any/At least one of Like/Contains (if not specified auto wraps the right string OPERAND in a "%" for wildcard
match) ?!~ Any/At least one of NOT Like/Contains (if not specified auto wraps the right string OPERAND in a "%" for
wildcard match) To group and combine several expressions you can use parenthesis
(...), && (AND) and || (OR) tokens. Single line comments are also supported: // Example comment.
- **fields** (String): Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
API Logs    List logs    Returns a paginated logs list.
Only superusers can perform this action.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
const pageResult = await pb.logs.getList(1, 20, {
filter: 'data.status >= 400'
});  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
final pageResult = await pb.logs.getList(
page: 1,
perPage: 20,
filter: 'data.status >= 400',
);   ###### API details
GET /api/logs Requires `Authorization:TOKEN` Query parameters Param Type Description page Number The page (aka. offset) of the paginated list (default to 1). perPage Number The max returned logs per page (default to 30). sort String Specify the ORDER BY fields.
Add - / + (default) in front of the attribute for DESC /
ASC order, e.g.:
// DESC by the insertion rowid and ASC by level
?sort=-rowid,level  Supported log sort fields:  @random, rowid, id, created,
updated, level, message and any
data.* attribute.
filter String Filter expression to filter/search the returned logs list, e.g.:
`?filter=(data.url~'test.com' &amp;&amp; level>0)`  Supported log filter fields:  id, created, updated,
level, message and any data.* attribute.
The syntax basically follows the format
OPERAND OPERATOR OPERAND, where:
-OPERAND - could be any field literal, string (single or double quoted),
number, null, true, false
-OPERATOR - is one of:
= Equal
-!= NOT equal
-> Greater than
->= Greater than or equal
-&lt; Less than
-&lt;= Less than or equal
-~ Like/Contains (if not specified auto wraps the right string OPERAND in a &quot;%&quot; for wildcard
match)
-!~ NOT Like/Contains (if not specified auto wraps the right string OPERAND in a &quot;%&quot; for
wildcard match)
-?= Any/At least one of Equal
-?!= Any/At least one of NOT equal
-?> Any/At least one of Greater than
-?>= Any/At least one of Greater than or equal
-?&lt; Any/At least one of Less than
-?&lt;= Any/At least one of Less than or equal
-?~ Any/At least one of Like/Contains (if not specified auto wraps the right string OPERAND in a &quot;%&quot; for wildcard
match)
-?!~ Any/At least one of NOT Like/Contains (if not specified auto wraps the right string OPERAND in a &quot;%&quot; for
wildcard match)
To group and combine several expressions you can use parenthesis
(...), &amp;&amp; (AND) and || (OR) tokens.
Single line comments are also supported: // Example comment.
fields String Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name
* targets all keys from the specific depth level.
In addition, the following field modifiers are also supported:
-:excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
Responses  {
"page": 1,
"perPage": 20,
"totalItems": 2,
"items": [
"id": "ai5z3aoed6809au",
"created": "2024-10-27 09:28:19.524Z",
"data": {
"auth": "_superusers",
"execTime": 2.392327,
"method": "GET",
"remoteIP": "127.0.0.1",
"status": 200,
"type": "request",
"url": "/api/collections/_pbc_2287844090/records?page=1&amp;perPage=1&amp;filter=&amp;fields=id",
"userAgent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36",
"userIP": "127.0.0.1"
"message": "GET /api/collections/_pbc_2287844090/records?page=1&amp;perPage=1&amp;filter=&amp;fields=id",
"level": 0
"id": "26apis4s3sm9yqm",
"created": "2024-10-27 09:28:19.524Z",
"data": {
"auth": "_superusers",
"execTime": 2.392327,
"method": "GET",
"remoteIP": "127.0.0.1",
"status": 200,
"type": "request",
"url": "/api/collections/_pbc_2287844090/records?page=1&amp;perPage=1&amp;filter=&amp;fields=id",
"userAgent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36",
"userIP": "127.0.0.1"
"message": "GET /api/collections/_pbc_2287844090/records?page=1&amp;perPage=1&amp;filter=&amp;fields=id",
"level": 0
}  {
"status": 400,
"message": "Something went wrong while processing your request. Invalid filter.",
"data": {}
}  {
"status": 401,
"message": "The request requires valid record authorization token.",
"data": {}
}  {
"status": 403,
"message": "The authorized record is not allowed to perform this action.",
"data": {}
}      View log    Returns a single log by its ID.
Only superusers can perform this action.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
const log = await pb.logs.getOne('LOG_ID');  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
final log = await pb.logs.getOne('LOG_ID');   ###### API details
GET /api/logs/`id` Requires `Authorization:TOKEN` Path parameters Param Type Description id String ID of the log to view. Query parameters Param Type Description fields String Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name
* targets all keys from the specific depth level.
In addition, the following field modifiers are also supported:
-:excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
Responses  {
"id": "ai5z3aoed6809au",
"created": "2024-10-27 09:28:19.524Z",
"data": {
"auth": "_superusers",
"execTime": 2.392327,
"method": "GET",
"remoteIP": "127.0.0.1",
"status": 200,
"type": "request",
"url": "/api/collections/_pbc_2287844090/records?page=1&amp;perPage=1&amp;filter=&amp;fields=id",
"userAgent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36",
"userIP": "127.0.0.1"
"message": "GET /api/collections/_pbc_2287844090/records?page=1&amp;perPage=1&amp;filter=&amp;fields=id",
"level": 0
}  {
"status": 401,
"message": "The request requires valid record authorization token.",
"data": {}
}  {
"status": 403,
"message": "The authorized record is not allowed to perform this action.",
"data": {}
}  {
"status": 404,
"message": "The requested resource wasn't found.",
"data": {}
}      Logs statistics    Returns hourly aggregated logs statistics.
Only superusers can perform this action.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
const stats = await pb.logs.getStats({
filter: 'data.status >= 400'
});  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
final stats = await pb.logs.getStats(
filter: 'data.status >= 400'
);   ###### API details
GET /api/logs/stats Requires `Authorization:TOKEN` Query parameters Param Type Description filter String Filter expression to filter/search the logs, e.g.:
`?filter=(data.url~'test.com' &amp;&amp; level>0)`  Supported log filter fields:  rowid, id, created,
updated, level, message and any
data.* attribute.
The syntax basically follows the format
OPERAND OPERATOR OPERAND, where:
-OPERAND - could be any field literal, string (single or double quoted),
number, null, true, false
-OPERATOR - is one of:
= Equal
-!= NOT equal
-> Greater than
->= Greater than or equal
-&lt; Less than
-&lt;= Less than or equal
-~ Like/Contains (if not specified auto wraps the right string OPERAND in a &quot;%&quot; for wildcard
match)
-!~ NOT Like/Contains (if not specified auto wraps the right string OPERAND in a &quot;%&quot; for
wildcard match)
-?= Any/At least one of Equal
-?!= Any/At least one of NOT equal
-?> Any/At least one of Greater than
-?>= Any/At least one of Greater than or equal
-?&lt; Any/At least one of Less than
-?&lt;= Any/At least one of Less than or equal
-?~ Any/At least one of Like/Contains (if not specified auto wraps the right string OPERAND in a &quot;%&quot; for wildcard
match)
-?!~ Any/At least one of NOT Like/Contains (if not specified auto wraps the right string OPERAND in a &quot;%&quot; for
wildcard match)
To group and combine several expressions you can use parenthesis
(...), &amp;&amp; (AND) and || (OR) tokens.
Single line comments are also supported: // Example comment.
fields String Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name
* targets all keys from the specific depth level.
In addition, the following field modifiers are also supported:
-:excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
Responses  [
"total": 4,
"date": "2022-06-01 19:00:00.000"
"total": 1,
"date": "2022-06-02 12:00:00.000"
"total": 8,
"date": "2022-06-02 13:00:00.000"
]  {
"status": 400,
"message": "Something went wrong while processing your request. Invalid filter.",
"data": {}
}  {
"status": 401,
"message": "The request requires valid record authorization token.",
"data": {}
}  {
"status": 403,
"message": "The authorized record is not allowed to perform this action.",
"data": {}

## 23.Web APIs reference - API Health
fields(String):Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
# Web APIs reference - API Health
- **fields** (String): Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
API Health    Health check    Returns the health status of the server.
###### API details
GET/HEAD /api/health Query parameters Param Type Description fields String Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name
* targets all keys from the specific depth level.
In addition, the following field modifiers are also supported:
-:excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
Responses  {
"status": 200,
"message": "API is healthy.",
"data": {
"canBackup": false

## 24.Web APIs reference - API Backups
fields(String):Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
Optional name(String):The base name of the backup file to create.
Must be in the format [a-z0-9_-].zip
If not set, it will be auto generated.
Required file(File):The zip archive to upload.
key(String):The key of the backup file to delete.
key(String):The key of the backup file to restore.
key(String):The key of the backup file to download.
token(String):Superuser file token for granting access to the
backup file.
# Web APIs reference - API Backups
- **fields** (String): Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name * targets all keys from the specific depth level. In addition, the following field modifiers are also supported: :excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
- **Optional name** (String): The base name of the backup file to create.
If not set, it will be auto generated.
- **Required file** (File): The zip archive to upload.
- **key** (String): The key of the backup file to delete.
- **key** (String): The key of the backup file to restore.
- **token** (String): Superuser file token for granting access to the
backup file.
API Backups    List backups    Returns list with all available backup files.
Only superusers can perform this action.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
const backups = await pb.backups.getFullList();  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
final backups = await pb.backups.getFullList();   ###### API details
GET /api/backups Query parameters Param Type Description fields String Comma separated string of the fields to return in the JSON response
(by default returns all fields). Ex.:
?fields=*,expand.relField.name
* targets all keys from the specific depth level.
In addition, the following field modifiers are also supported:
-:excerpt(maxLength, withEllipsis?)
Returns a short plain text version of the field string value.
Ex.:
?fields=*,description:excerpt(200,true)
Responses  [
"modified": "2023-05-19 16:25:57.542Z",
"size": 251316185
"modified": "2023-05-18 16:25:57.542Z",
"size": 251314010
]  {
"status": 400,
"message": "Failed to load backups filesystem.",
"data": {}
}  {
"status": 401,
"message": "The request requires valid record authorization token.",
"data": {}
}  {
"status": 403,
"message": "Only superusers can perform this action.",
"data": {}
}      Create backup    Creates a new app data backup.
This action will return an error if there is another backup/restore operation already in progress.
Only superusers can perform this action.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
final pb = PocketBase('http://127.0.0.1:8090');
POST /api/backups Requires `Authorization:TOKEN` Body Parameters Param Type Description Optional name String The base name of the backup file to create.
If not set, it will be auto generated. Body parameters could be sent as JSON or
multipart/form-data. Responses  `null`  {
"status": 400,
"message": "Try again later - another backup/restore process has already been started.",
"data": {}
}  {
"status": 401,
"message": "The request requires valid record authorization token.",
"data": {}
}  {
"status": 403,
"message": "The authorized record is not allowed to perform this action.",
"data": {}
}      Upload backup    Uploads an existing backup zip file.
Only superusers can perform this action.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
await pb.backups.upload({ file: new Blob([...]) });  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
await pb.backups.upload(http.MultipartFile.fromBytes('file', ...));   ###### API details
POST /api/backups/upload Requires `Authorization:TOKEN` Body Parameters Param Type Description Required file File The zip archive to upload. Uploading files is supported only via multipart/form-data. Responses  `null`  {
"status": 400,
"message": "Something went wrong while processing your request.",
"data": {
"file": {
"code": "validation_invalid_mime_type",
"message": "\"test_backup.txt\" mime type must be one of: application/zip."
}  {
"status": 401,
"message": "The request requires valid record authorization token.",
"data": {}
}  {
"status": 403,
"message": "The authorized record is not allowed to perform this action.",
"data": {}
}      Delete backup    Deletes a single backup by its name.
This action will return an error if the backup to delete is still being generated or part of a
restore operation.
Only superusers can perform this action.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
final pb = PocketBase('http://127.0.0.1:8090');
DELETE /api/backups/`key` Requires `Authorization:TOKEN` Path parameters Param Type Description key String The key of the backup file to delete. Responses  `null`  {
"status": 400,
"message": "Try again later - another backup/restore process has already been started.",
"data": {}
}  {
"status": 401,
"message": "The request requires valid record authorization token.",
"data": {}
}  {
"status": 403,
"message": "The authorized record is not allowed to perform this action.",
"data": {}
}      Restore backup    Restore a single backup by its name and restarts the current running PocketBase process.
This action will return an error if there is another backup/restore operation already in progress.
Only superusers can perform this action.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
final pb = PocketBase('http://127.0.0.1:8090');
POST /api/backups/`key`/restore Requires `Authorization:TOKEN` Path parameters Param Type Description key String The key of the backup file to restore. Responses  `null`  {
"status": 400,
"message": "Try again later - another backup/restore process has already been started.",
"data": {}
}  {
"status": 401,
"message": "The request requires valid record authorization token.",
"data": {}
}  {
"status": 403,
"message": "The authorized record is not allowed to perform this action.",
"data": {}
Only superusers can perform this action.
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
const token = await pb.files.getToken();
final pb = PocketBase('http://127.0.0.1:8090');
final token = await pb.files.getToken();
backup file. Responses  `[file resource]`  {
"status": 400,
"message": "Filesystem initialization failure.",
"data": {}
}  {
"status": 404,
"message": "The requested resource wasn't found.",
"data": {}

## 25.Extend with Go - Overview
# Extend with Go - Overview
Overview  ### Getting started
PocketBase can be used as regular Go package that exposes various helpers and hooks to help you implement
your own custom portable application.
A new PocketBase instance is created via
pocketbase.New()
pocketbase.NewWithConfig(config)
Once created you can register your custom business logic via the available
event hooks
and call
app.Start()
to start the application.
Below is a minimal example:
-Install Go 1.23+
-Create a new project directory with main.go file inside it.
As a reference, you can also explore the prebuilt executable
example/base/main.go
file.
package main
import (
"log"
"os"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/apis"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
app.OnServe().BindFunc(func(se *core.ServeEvent) error {
// serves static files from the provided public dir (if exists)
se.Router.GET("/{path...}", apis.Static(os.DirFS("./pb_public"), false))
if err := app.Start(); err != nil {
log.Fatal(err)
-To init the dependencies, run go mod init myapp &amp;&amp; go mod tidy.
-To start the application, run go run . serve.
-To build a statically linked executable, run go build
and then you can start the created executable with
./myapp serve.
### Custom SQLite driver
The general recommendation is to use the builtin SQLite setup but if you need more
advanced configuration or extensions like ICU, FTS5, etc. you&#39;ll have to specify a custom driver/build.
Note that PocketBase by default doesn&#39;t require CGO because it uses the pure Go SQLite port
modernc.org/sqlite
, but this may not be the case when using a custom SQLite driver!
PocketBase v0.23+ added support for defining a DBConnect function as app configuration to
load custom SQLite builds and drivers compatible with the standard Go database/sql.
The DBConnect function is called twice - once for
pb_data/data.db
(the main database file) and second time for pb_data/auxiliary.db (used for logs and other ephemeral
system meta information).
If you want to load your custom driver conditionally and fallback to the default handler, then you can
call
core.DefaultDBConnect
As a side-note, if you are not planning to use core.DefaultDBConnect
fallback as part of your custom driver registration you can exclude the default pure Go driver with
go build -tags no_default_driver to reduce the binary size a little (~4MB).
Below are some minimal examples with commonly used external SQLite drivers:
github.com/mattn/go-sqlite3    For all available options please refer to the
github.com/mattn/go-sqlite3
README.
package main
import (
"database/sql"
"log"
"github.com/mattn/go-sqlite3"
"github.com/pocketbase/dbx"
"github.com/pocketbase/pocketbase"
// register a new driver with default PRAGMAs and the same query
// builder implementation as the already existing sqlite3 builder
func init() {
// initialize default PRAGMAs for each new connection
sql.Register("pb_sqlite3",
&amp;sqlite3.SQLiteDriver{
ConnectHook: func(conn *sqlite3.SQLiteConn) error {
_, err := conn.Exec(`
PRAGMA busy_timeout       = 10000;
PRAGMA journal_mode       = WAL;
PRAGMA journal_size_limit = 200000000;
PRAGMA synchronous        = NORMAL;
PRAGMA foreign_keys       = ON;
PRAGMA temp_store         = MEMORY;
PRAGMA cache_size         = -16000;
`, nil)
return err
dbx.BuilderFuncMap["pb_sqlite3"] = dbx.BuilderFuncMap["sqlite3"]
func main() {
app := pocketbase.NewWithConfig(pocketbase.Config{
DBConnect: func(dbPath string) (*dbx.DB, error) {
return dbx.Open("pb_sqlite3", dbPath)
// any custom hooks or plugins...
if err := app.Start(); err != nil {
log.Fatal(err)
}     github.com/ncruces/go-sqlite3    For all available options please refer to the
github.com/ncruces/go-sqlite3
README.
package main
import (
"log"
"github.com/pocketbase/dbx"
"github.com/pocketbase/pocketbase"
_ "github.com/ncruces/go-sqlite3/driver"
_ "github.com/ncruces/go-sqlite3/embed"
func main() {
app := pocketbase.NewWithConfig(pocketbase.Config{
DBConnect: func(dbPath string) (*dbx.DB, error) {
const pragmas = "?_pragma=busy_timeout(10000)&amp;_pragma=journal_mode(WAL)&amp;_pragma=journal_size_limit(200000000)&amp;_pragma=synchronous(NORMAL)&amp;_pragma=foreign_keys(ON)&amp;_pragma=temp_store(MEMORY)&amp;_pragma=cache_size(-16000)"
return dbx.Open("sqlite3", "file:"+dbPath+pragmas)
// custom hooks and plugins...
if err := app.Start(); err != nil {
log.Fatal(err)

## 26.Extend with Go - Event hooks
GET /hello"
# Extend with Go - Event hooks
`GET /hello"`
Event hooks The standard way to modify PocketBase is through
event hooks in your Go code.
All hooks have 3 main methods:
-Bind(handler)
adds a new handler to the specified event hook. A handler has 3 fields:
Id (optional) - the name of the handler (could be used
as argument for Unbind)
-Priority (optional) - the execution order of the handler
(if empty fallbacks to the order of registration in the code).
-Func (required) - the handler function.
-BindFunc(func)
is similar to Bind but registers a new handler from just the provided function.
The registered handler is added with a default 0 priority and the id is autogenerated (the returned string
value).
-Trigger(event, oneOffHandlerFuncs...)
triggers the event hook.
This method rarely has to be called manually by users.
To remove an already registered hook handler, you can use the handler id and pass it to
Unbind(id) or remove all handlers with
UnbindAll() (!including system handlers).
All hook handler functions share the same func(e T) error signature and expect
If you need to access the app instance from inside a hook handler, prefer using the
e.App field instead of reusing a parent scope app variable because the hook could be part
of a DB transaction and can cause deadlock.
Also avoid using global mutex locks inside a hook handler because it could be invoked recursively
(e.g. cascade delete) and can cause deadlock.
You can explore all available hooks below:
### App hooks
OnBootstrap
OnBootstrap hook is triggered when initializing the main
application resources (db, app settings, etc).
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
return err
// e.App
return nil
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnServe
`OnServe` hook is triggered when the app web server is started
(after starting the TCP listener but before initializing the blocking serve task),
allowing you to adjust its options and attach new routes or middlewares.
package main
import (
"log"
"net/http"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/apis"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
app.OnServe().BindFunc(func(e *core.ServeEvent) error {
// register new "GET /hello" route
e.Router.GET("/hello", func(e *core.RequestEvent) error {
return e.String(200, "Hello world!")
}).Bind(apis.RequireAuth())
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnSettingsReload
OnSettingsReload hook is triggered every time when the App.Settings()
is being replaced with a new state.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
app.OnSettingsReload().BindFunc(func(e *core.SettingsReloadEvent) error {
return err
// e.App.Settings()
return nil
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnBackupCreate
`OnBackupCreate` is triggered on each `App.CreateBackup` call.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
app.OnBackupCreate().BindFunc(func(e *core.BackupEvent) error {
// e.App
// e.Name    - the name of the backup to create
// e.Exclude - list of pb_data dir entries to exclude from the backup
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnBackupRestore
`OnBackupRestore` is triggered before app backup restore (aka. on `App.RestoreBackup` call).
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
app.OnBackupRestore().BindFunc(func(e *core.BackupEvent) error {
// e.App
// e.Name    - the name of the backup to restore
// e.Exclude - list of dir entries to exclude from the backup
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnTerminate
`OnTerminate` hook is triggered when the app is in the process
of being terminated (ex. on `SIGTERM` signal).
Note that the app could be terminated abruptly without awaiting the hook completion.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
app.OnTerminate().BindFunc(func(e *core.TerminateEvent) error {
// e.App
// e.IsRestart
if err := app.Start(); err != nil {
log.Fatal(err)
}   ### Mailer hooks
OnMailerSend
OnMailerSend hook is triggered every time when a new email is
being send using the App.NewMailClient() instance.
It allows intercepting the email message or to use a custom mailer client.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
app.OnMailerSend().BindFunc(func(e *core.MailerEvent) error {
// e.App
// e.Mailer
// e.Message
// ex. change the mail subject
e.Message.Subject = "new subject"
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnMailerRecordAuthAlertSend
OnMailerRecordAuthAlertSend hook is triggered when
sending a new device login auth alert email, allowing you to
intercept and customize the email message that is being sent.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
app.OnMailerRecordAuthAlertSend().BindFunc(func(e *core.MailerRecordEvent) error {
// e.App
// e.Mailer
// e.Message
// e.Record
// e.Meta
// ex. change the mail subject
e.Message.Subject = "new subject"
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnMailerRecordPasswordResetSend
OnMailerRecordPasswordResetSend hook is triggered when
sending a password reset email to an auth record, allowing
you to intercept and customize the email message that is being sent.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
app.OnMailerRecordPasswordResetSend().BindFunc(func(e *core.MailerRecordEvent) error {
// e.App
// e.Mailer
// e.Message
// e.Record
// e.Meta
// ex. change the mail subject
e.Message.Subject = "new subject"
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnMailerRecordVerificationSend
OnMailerRecordVerificationSend hook is triggered when
sending a verification email to an auth record, allowing
you to intercept and customize the email message that is being sent.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
app.OnMailerRecordVerificationSend().BindFunc(func(e *core.MailerRecordEvent) error {
// e.App
// e.Mailer
// e.Message
// e.Record
// e.Meta
// ex. change the mail subject
e.Message.Subject = "new subject"
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnMailerRecordEmailChangeSend
OnMailerRecordEmailChangeSend hook is triggered when sending a
confirmation new address email to an auth record, allowing
you to intercept and customize the email message that is being sent.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
app.OnMailerRecordEmailChangeSend().BindFunc(func(e *core.MailerRecordEvent) error {
// e.App
// e.Mailer
// e.Message
// e.Record
// e.Meta
// ex. change the mail subject
e.Message.Subject = "new subject"
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnMailerRecordOTPSend
OnMailerRecordOTPSend hook is triggered when sending an OTP email
to an auth record, allowing you to intercept and customize the
email message that is being sent.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
app.OnMailerRecordOTPSend().BindFunc(func(e *core.MailerRecordEvent) error {
// e.App
// e.Mailer
// e.Message
// e.Record
// e.Meta
// ex. change the mail subject
e.Message.Subject = "new subject"
if err := app.Start(); err != nil {
log.Fatal(err)
}   ### Realtime hooks
OnRealtimeConnectRequest
OnRealtimeConnectRequest hook is triggered when establishing the SSE client connection.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
app.OnRealtimeConnectRequest().BindFunc(func(e *core.RealtimeConnectRequestEvent) error {
// e.App
// e.Client
// e.IdleTimeout
// and all RequestEvent fields...
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnRealtimeSubscribeRequest
`OnRealtimeSubscribeRequest` hook is triggered when updating the
client subscriptions, allowing you to further validate and
modify the submitted change.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
app.OnRealtimeSubscribeRequest().BindFunc(func(e *core.RealtimeSubscribeRequestEvent) error {
// e.App
// e.Client
// e.Subscriptions
// and all RequestEvent fields...
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnRealtimeMessageSend
`OnRealtimeMessageSend` hook is triggered when sending an SSE message to a client.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
app.OnRealtimeMessageSend().BindFunc(func(e *core.RealtimeMessageEvent) error {
// e.App
// e.Client
// e.Message
// and all original connect RequestEvent fields...
if err := app.Start(); err != nil {
log.Fatal(err)
}   ### Record model hooks
These are lower level Record model hooks and could be triggered from anywhere (custom console command, scheduled cron job, when calling e.Save(record), etc.) and therefore they have no access to the request context!
If you want to intercept the builtin Web APIs and to access their request body, query parameters, headers or the request auth state, then please use the designated
Record *Request hooks
OnRecordEnrich
OnRecordEnrich is triggered every time when a record is enriched
- as part of the builtin Record responses, during realtime message serialization, or when apis.EnrichRecord is invoked.
It could be used for example to redact/hide or add computed temporary
Record model props only for the specific request info.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
app.OnRecordEnrich("posts").BindFunc(func(e *core.RecordEnrichEvent) error {
// hide one or more fields
e.Record.Hide("role")
// add new custom field for registered users
if e.RequestInfo.Auth != nil &amp;&amp; e.RequestInfo.Auth.Collection().Name == "users" {
e.Record.WithCustomData(true) // for security requires explicitly allowing it
e.Record.Set("computedScore", e.Record.GetInt("score") * e.RequestInfo.Auth.GetInt("base"))
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnRecordValidate
OnRecordValidate is a Record proxy model hook of OnModelValidate.
OnRecordValidate is called every time when a Record is being validated,
e.g. triggered by App.Validate() or App.Save().
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every record
app.OnRecordValidate().BindFunc(func(e *core.RecordEvent) error {
// e.App
// e.Record
// fires only for "users" and "articles" records
app.OnRecordValidate("users", "articles").BindFunc(func(e *core.RecordEvent) error {
// e.App
// e.Record
if err := app.Start(); err != nil {
log.Fatal(err)
}   ###### Record model create hooks
OnRecordCreate
OnRecordCreate is a Record proxy model hook of OnModelCreate.
OnRecordCreate is triggered every time when a new Record is being created,
e.g. triggered by App.Save().
and the INSERT DB statement.
and the INSERT DB statement.
Note that successful execution doesn't guarantee that the Record
is persisted in the database since its wrapping transaction may
not have been committed yet.
If you want to listen to only the actual persisted events, you can
bind to `OnRecordAfterCreateSuccess` or `OnRecordAfterCreateError` hooks.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every record
app.OnRecordCreate().BindFunc(func(e *core.RecordEvent) error {
// e.App
// e.Record
// fires only for "users" and "articles" records
app.OnRecordCreate("users", "articles").BindFunc(func(e *core.RecordEvent) error {
// e.App
// e.Record
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnRecordCreateExecute
OnRecordCreateExecute is a Record proxy model hook of OnModelCreateExecute.
OnRecordCreateExecute is triggered after successful Record validation
and right before the model INSERT DB statement execution.
Usually it is triggered as part of the App.Save() in the following firing order:
OnRecordCreate
&nbsp;->
OnRecordValidate (skipped with App.SaveNoValidate())
&nbsp;->
OnRecordCreateExecute
Note that successful execution doesn't guarantee that the Record
is persisted in the database since its wrapping transaction may
not have been committed yet.
If you want to listen to only the actual persisted events, you can
bind to `OnRecordAfterCreateSuccess` or `OnRecordAfterCreateError` hooks.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every record
app.OnRecordCreateExecute().BindFunc(func(e *core.RecordEvent) error {
// e.App
// e.Record
// fires only for "users" and "articles" records
app.OnRecordCreateExecute("users", "articles").BindFunc(func(e *core.RecordEvent) error {
// e.App
// e.Record
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnRecordAfterCreateSuccess
OnRecordAfterCreateSuccess is a Record proxy model hook of OnModelAfterCreateSuccess.
OnRecordAfterCreateSuccess is triggered after each successful
Record DB create persistence.
Note that when a Record is persisted as part of a transaction,
this hook is delayed and executed only AFTER the transaction has been committed.
This hook is NOT triggered in case the transaction fails/rollbacks.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every record
app.OnRecordAfterCreateSuccess().BindFunc(func(e *core.RecordEvent) error {
// e.App
// e.Record
// fires only for "users" and "articles" records
app.OnRecordAfterCreateSuccess("users", "articles").BindFunc(func(e *core.RecordEvent) error {
// e.App
// e.Record
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnRecordAfterCreateError
OnRecordAfterCreateError is a Record proxy model hook of OnModelAfterCreateError.
OnRecordAfterCreateError is triggered after each failed
Record DB create persistence.
Note that the execution of this hook is either immediate or delayed
depending on the error:
-immediate on App.Save() failure
-delayed on transaction rollback
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every record
app.OnRecordAfterCreateError().BindFunc(func(e *core.RecordErrorEvent) error {
// e.App
// e.Record
// e.Error
// fires only for "users" and "articles" records
app.OnRecordAfterCreateError("users", "articles").BindFunc(func(e *core.RecordErrorEvent) error {
// e.App
// e.Record
// e.Error
if err := app.Start(); err != nil {
log.Fatal(err)
}   ###### Record model update hooks
OnRecordUpdate
OnRecordUpdate is a Record proxy model hook of OnModelUpdate.
OnRecordUpdate is triggered every time when a new Record is being updated,
e.g. triggered by App.Save().
and the UPDATE DB statement.
and the UPDATE DB statement.
Note that successful execution doesn't guarantee that the Record
is persisted in the database since its wrapping transaction may
not have been committed yet.
If you want to listen to only the actual persisted events, you can
bind to `OnRecordAfterUpdateSuccess` or `OnRecordAfterUpdateError` hooks.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every record
app.OnRecordUpdate().BindFunc(func(e *core.RecordEvent) error {
// e.App
// e.Record
// fires only for "users" and "articles" records
app.OnRecordUpdate("users", "articles").BindFunc(func(e *core.RecordEvent) error {
// e.App
// e.Record
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnRecordUpdateExecute
OnRecordUpdateExecute is a Record proxy model hook of OnModelUpdateExecute.
OnRecordUpdateExecute is triggered after successful Record validation
and right before the model UPDATE DB statement execution.
Usually it is triggered as part of the App.Save() in the following firing order:
OnRecordUpdate
&nbsp;->
OnRecordValidate (skipped with App.SaveNoValidate())
&nbsp;->
OnRecordUpdateExecute
Note that successful execution doesn't guarantee that the Record
is persisted in the database since its wrapping transaction may
not have been committed yet.
If you want to listen to only the actual persisted events, you can
bind to `OnRecordAfterUpdateSuccess` or `OnRecordAfterUpdateError` hooks.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every record
app.OnRecordUpdateExecute().BindFunc(func(e *core.RecordEvent) error {
// e.App
// e.Record
// fires only for "users" and "articles" records
app.OnRecordUpdateExecute("users", "articles").BindFunc(func(e *core.RecordEvent) error {
// e.App
// e.Record
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnRecordAfterUpdateSuccess
OnRecordAfterUpdateSuccess is a Record proxy model hook of OnModelAfterUpdateSuccess.
OnRecordAfterUpdateSuccess is triggered after each successful
Record DB update persistence.
Note that when a Record is persisted as part of a transaction,
this hook is delayed and executed only AFTER the transaction has been committed.
This hook is NOT triggered in case the transaction fails/rollbacks.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every record
app.OnRecordAfterUpdateSuccess().BindFunc(func(e *core.RecordEvent) error {
// e.App
// e.Record
// fires only for "users" and "articles" records
app.OnRecordAfterUpdateSuccess("users", "articles").BindFunc(func(e *core.RecordEvent) error {
// e.App
// e.Record
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnRecordAfterUpdateError
OnRecordAfterUpdateError is a Record proxy model hook of OnModelAfterUpdateError.
OnRecordAfterUpdateError is triggered after each failed
Record DB update persistence.
Note that the execution of this hook is either immediate or delayed
depending on the error:
-immediate on App.Save() failure
-delayed on transaction rollback
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every record
app.OnRecordAfterUpdateError().BindFunc(func(e *core.RecordErrorEvent) error {
// e.App
// e.Record
// e.Error
// fires only for "users" and "articles" records
app.OnRecordAfterUpdateError("users", "articles").BindFunc(func(e *core.RecordErrorEvent) error {
// e.App
// e.Record
// e.Error
if err := app.Start(); err != nil {
log.Fatal(err)
}   ###### Record model delete hooks
OnRecordDelete
OnRecordDelete is a Record proxy model hook of OnModelDelete.
OnRecordDelete is triggered every time when a new Record is being deleted,
e.g. triggered by App.Delete().
and the UPDATE DB statement.
and the UPDATE DB statement.
Note that successful execution doesn't guarantee that the Record
is deleted from the database since its wrapping transaction may
not have been committed yet.
If you want to listen to only the actual persisted deleted events, you can
bind to `OnRecordAfterDeleteSuccess` or `OnRecordAfterDeleteError` hooks.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every record
app.OnRecordDelete().BindFunc(func(e *core.RecordEvent) error {
// e.App
// e.Record
// fires only for "users" and "articles" records
app.OnRecordDelete("users", "articles").BindFunc(func(e *core.RecordEvent) error {
// e.App
// e.Record
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnRecordDeleteExecute
OnRecordDeleteExecute is a Record proxy model hook of OnModelDeleteExecute.
OnRecordDeleteExecute is triggered after the internal delete checks and
right before the Record the model DELETE DB statement execution.
Usually it is triggered as part of the App.Delete() in the following firing order:
OnRecordDelete
&nbsp;->
internal delete checks
&nbsp;->
OnRecordDeleteExecute
Note that successful execution doesn't guarantee that the Record
is deleted from the database since its wrapping transaction may
not have been committed yet.
If you want to listen to only the actual persisted events, you can
bind to `OnRecordAfterDeleteSuccess` or `OnRecordAfterDeleteError` hooks.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every record
app.OnRecordDeleteExecute().BindFunc(func(e *core.RecordEvent) error {
// e.App
// e.Record
// fires only for "users" and "articles" records
app.OnRecordDeleteExecute("users", "articles").BindFunc(func(e *core.RecordEvent) error {
// e.App
// e.Record
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnRecordAfterDeleteSuccess
OnRecordAfterDeleteSuccess is a Record proxy model hook of OnModelAfterDeleteSuccess.
OnRecordAfterDeleteSuccess is triggered after each successful
Record DB delete persistence.
Note that when a Record is deleted as part of a transaction,
this hook is delayed and executed only AFTER the transaction has been committed.
This hook is NOT triggered in case the transaction fails/rollbacks.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every record
app.OnRecordAfterDeleteSuccess().BindFunc(func(e *core.RecordEvent) error {
// e.App
// e.Record
// fires only for "users" and "articles" records
app.OnRecordAfterDeleteSuccess("users", "articles").BindFunc(func(e *core.RecordEvent) error {
// e.App
// e.Record
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnRecordAfterDeleteError
OnRecordAfterDeleteError is a Record proxy model hook of OnModelAfterDeleteError.
OnRecordAfterDeleteError is triggered after each failed
Record DB delete persistence.
Note that the execution of this hook is either immediate or delayed
depending on the error:
-immediate on App.Delete() failure
-delayed on transaction rollback
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every record
app.OnRecordAfterDeleteError().BindFunc(func(e *core.RecordErrorEvent) error {
// e.App
// e.Record
// e.Error
// fires only for "users" and "articles" records
app.OnRecordAfterDeleteError("users", "articles").BindFunc(func(e *core.RecordErrorEvent) error {
// e.App
// e.Record
// e.Error
if err := app.Start(); err != nil {
log.Fatal(err)
}   ### Collection model hooks
These are lower level Collection model hooks and could be triggered from anywhere (custom console command, scheduled cron job, when calling e.Save(collection), etc.) and therefore they have no access to the request context!
If you want to intercept the builtin Web APIs and to access their request body, query parameters, headers or the request auth state, then please use the designated
Collection *Request hooks
OnCollectionValidate
OnCollectionValidate is a Collection proxy model hook of OnModelValidate.
OnCollectionValidate is called every time when a Collection is being validated,
e.g. triggered by App.Validate() or App.Save().
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every collection
app.OnCollectionValidate().BindFunc(func(e *core.CollectionEvent) error {
// e.App
// e.Collection
// fires only for "users" and "articles" collections
app.OnCollectionValidate("users", "articles").BindFunc(func(e *core.CollectionEvent) error {
// e.App
// e.Collection
if err := app.Start(); err != nil {
log.Fatal(err)
}   ###### Collection mode create hooks
OnCollectionCreate
OnCollectionCreate is a Collection proxy model hook of OnModelCreate.
OnCollectionCreate is triggered every time when a new Collection is being created,
e.g. triggered by App.Save().
and the INSERT DB statement.
and the INSERT DB statement.
Note that successful execution doesn't guarantee that the Collection
is persisted in the database since its wrapping transaction may
not have been committed yet.
If you want to listen to only the actual persisted events, you can
bind to `OnCollectionAfterCreateSuccess` or `OnCollectionAfterCreateError` hooks.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every collection
app.OnCollectionCreate().BindFunc(func(e *core.CollectionEvent) error {
// e.App
// e.Collection
// fires only for "users" and "articles" collections
app.OnCollectionCreate("users", "articles").BindFunc(func(e *core.CollectionEvent) error {
// e.App
// e.Collection
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnCollectionCreateExecute
OnCollectionCreateExecute is a Collection proxy model hook of OnModelCreateExecute.
OnCollectionCreateExecute is triggered after successful Collection validation
and right before the model INSERT DB statement execution.
Usually it is triggered as part of the App.Save() in the following firing order:
OnCollectionCreate
&nbsp;->
OnCollectionValidate (skipped with App.SaveNoValidate())
&nbsp;->
OnCollectionCreateExecute
Note that successful execution doesn't guarantee that the Collection
is persisted in the database since its wrapping transaction may
not have been committed yet.
If you want to listen to only the actual persisted events, you can
bind to `OnCollectionAfterCreateSuccess` or `OnCollectionAfterCreateError` hooks.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every collection
app.OnCollectionCreateExecute().BindFunc(func(e *core.CollectionEvent) error {
// e.App
// e.Collection
// fires only for "users" and "articles" collections
app.OnCollectionCreateExecute("users", "articles").BindFunc(func(e *core.CollectionEvent) error {
// e.App
// e.Collection
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnCollectionAfterCreateSuccess
OnCollectionAfterCreateSuccess is a Collection proxy model hook of OnModelAfterCreateSuccess.
OnCollectionAfterCreateSuccess is triggered after each successful
Collection DB create persistence.
Note that when a Collection is persisted as part of a transaction,
this hook is delayed and executed only AFTER the transaction has been committed.
This hook is NOT triggered in case the transaction fails/rollbacks.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every collection
app.OnCollectionAfterCreateSuccess().BindFunc(func(e *core.CollectionEvent) error {
// e.App
// e.Collection
// fires only for "users" and "articles" collections
app.OnCollectionAfterCreateSuccess("users", "articles").BindFunc(func(e *core.CollectionEvent) error {
// e.App
// e.Collection
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnCollectionAfterCreateError
OnCollectionAfterCreateError is a Collection proxy model hook of OnModelAfterCreateError.
OnCollectionAfterCreateError is triggered after each failed
Collection DB create persistence.
Note that the execution of this hook is either immediate or delayed
depending on the error:
-immediate on App.Save() failure
-delayed on transaction rollback
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every collection
app.OnCollectionAfterCreateError().BindFunc(func(e *core.CollectionErrorEvent) error {
// e.App
// e.Collection
// e.Error
// fires only for "users" and "articles" collections
app.OnCollectionAfterCreateError("users", "articles").BindFunc(func(e *core.CollectionErrorEvent) error {
// e.App
// e.Collection
// e.Error
if err := app.Start(); err != nil {
log.Fatal(err)
}   ###### Collection mode update hooks
OnCollectionUpdate
OnCollectionUpdate is a Collection proxy model hook of OnModelUpdate.
OnCollectionUpdate is triggered every time when a new Collection is being updated,
e.g. triggered by App.Save().
and the UPDATE DB statement.
and the UPDATE DB statement.
Note that successful execution doesn't guarantee that the Collection
is persisted in the database since its wrapping transaction may
not have been committed yet.
If you want to listen to only the actual persisted events, you can
bind to `OnCollectionAfterUpdateSuccess` or `OnCollectionAfterUpdateError` hooks.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every collection
app.OnCollectionUpdate().BindFunc(func(e *core.CollectionEvent) error {
// e.App
// e.Collection
// fires only for "users" and "articles" collections
app.OnCollectionUpdate("users", "articles").BindFunc(func(e *core.CollectionEvent) error {
// e.App
// e.Collection
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnCollectionUpdateExecute
OnCollectionUpdateExecute is a Collection proxy model hook of OnModelUpdateExecute.
OnCollectionUpdateExecute is triggered after successful Collection validation
and right before the model UPDATE DB statement execution.
Usually it is triggered as part of the App.Save() in the following firing order:
OnCollectionUpdate
&nbsp;->
OnCollectionValidate (skipped with App.SaveNoValidate())
&nbsp;->
OnCollectionUpdateExecute
Note that successful execution doesn't guarantee that the Collection
is persisted in the database since its wrapping transaction may
not have been committed yet.
If you want to listen to only the actual persisted events, you can
bind to `OnCollectionAfterUpdateSuccess` or `OnCollectionAfterUpdateError` hooks.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every collection
app.OnCollectionUpdateExecute().BindFunc(func(e *core.CollectionEvent) error {
// e.App
// e.Collection
// fires only for "users" and "articles" collections
app.OnCollectionUpdateExecute("users", "articles").BindFunc(func(e *core.CollectionEvent) error {
// e.App
// e.Collection
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnCollectionAfterUpdateSuccess
OnCollectionAfterUpdateSuccess is a Collection proxy model hook of OnModelAfterUpdateSuccess.
OnCollectionAfterUpdateSuccess is triggered after each successful
Collection DB update persistence.
Note that when a Collection is persisted as part of a transaction,
this hook is delayed and executed only AFTER the transaction has been committed.
This hook is NOT triggered in case the transaction fails/rollbacks.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every collection
app.OnCollectionAfterUpdateSuccess().BindFunc(func(e *core.CollectionEvent) error {
// e.App
// e.Collection
// fires only for "users" and "articles" collections
app.OnCollectionAfterUpdateSuccess("users", "articles").BindFunc(func(e *core.CollectionEvent) error {
// e.App
// e.Collection
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnCollectionAfterUpdateError
OnCollectionAfterUpdateError is a Collection proxy model hook of OnModelAfterUpdateError.
OnCollectionAfterUpdateError is triggered after each failed
Collection DB update persistence.
Note that the execution of this hook is either immediate or delayed
depending on the error:
-immediate on App.Save() failure
-delayed on transaction rollback
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every collection
app.OnCollectionAfterUpdateError().BindFunc(func(e *core.CollectionErrorEvent) error {
// e.App
// e.Collection
// e.Error
// fires only for "users" and "articles" collections
app.OnCollectionAfterUpdateError("users", "articles").BindFunc(func(e *core.CollectionErrorEvent) error {
// e.App
// e.Collection
// e.Error
if err := app.Start(); err != nil {
log.Fatal(err)
}   ###### Collection mode delete hooks
OnCollectionDelete
OnCollectionDelete is a Collection proxy model hook of OnModelDelete.
OnCollectionDelete is triggered every time when a new Collection is being deleted,
e.g. triggered by App.Delete().
and the UPDATE DB statement.
and the UPDATE DB statement.
Note that successful execution doesn't guarantee that the Collection
is deleted from the database since its wrapping transaction may
not have been committed yet.
If you want to listen to only the actual persisted deleted events, you can
bind to `OnCollectionAfterDeleteSuccess` or `OnCollectionAfterDeleteError` hooks.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every collection
app.OnCollectionDelete().BindFunc(func(e *core.CollectionEvent) error {
// e.App
// e.Collection
// fires only for "users" and "articles" collections
app.OnCollectionDelete("users", "articles").BindFunc(func(e *core.CollectionEvent) error {
// e.App
// e.Collection
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnCollectionDeleteExecute
OnCollectionDeleteExecute is a Collection proxy model hook of OnModelDeleteExecute.
OnCollectionDeleteExecute is triggered after the internal delete checks and
right before the Collection the model DELETE DB statement execution.
Usually it is triggered as part of the App.Delete() in the following firing order:
OnCollectionDelete
&nbsp;->
internal delete checks
&nbsp;->
OnCollectionDeleteExecute
Note that successful execution doesn't guarantee that the Collection
is deleted from the database since its wrapping transaction may
not have been committed yet.
If you want to listen to only the actual persisted events, you can
bind to `OnCollectionAfterDeleteSuccess` or `OnCollectionAfterDeleteError` hooks.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every collection
app.OnCollectionDeleteExecute().BindFunc(func(e *core.CollectionEvent) error {
// e.App
// e.Collection
// fires only for "users" and "articles" collections
app.OnCollectionDeleteExecute("users", "articles").BindFunc(func(e *core.CollectionEvent) error {
// e.App
// e.Collection
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnCollectionAfterDeleteSuccess
OnCollectionAfterDeleteSuccess is a Collection proxy model hook of OnModelAfterDeleteSuccess.
OnCollectionAfterDeleteSuccess is triggered after each successful
Collection DB delete persistence.
Note that when a Collection is deleted as part of a transaction,
this hook is delayed and executed only AFTER the transaction has been committed.
This hook is NOT triggered in case the transaction fails/rollbacks.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every collection
app.OnCollectionAfterDeleteSuccess().BindFunc(func(e *core.CollectionEvent) error {
// e.App
// e.Collection
// fires only for "users" and "articles" collections
app.OnCollectionAfterDeleteSuccess("users", "articles").BindFunc(func(e *core.CollectionEvent) error {
// e.App
// e.Collection
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnCollectionAfterDeleteError
OnCollectionAfterDeleteError is a Collection proxy model hook of OnModelAfterDeleteError.
OnCollectionAfterDeleteError is triggered after each failed
Collection DB delete persistence.
Note that the execution of this hook is either immediate or delayed
depending on the error:
-immediate on App.Delete() failure
-delayed on transaction rollback
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every collection
app.OnCollectionAfterDeleteError().BindFunc(func(e *core.CollectionErrorEvent) error {
// e.App
// e.Collection
// e.Error
// fires only for "users" and "articles" collections
app.OnCollectionAfterDeleteError("users", "articles").BindFunc(func(e *core.CollectionErrorEvent) error {
// e.App
// e.Collection
// e.Error
if err := app.Start(); err != nil {
log.Fatal(err)
}   ### Request hooks
The request hooks are triggered only when the corresponding API request endpoint is accessed.
###### Record CRUD request hooks
OnRecordsListRequest
OnRecordsListRequest hook is triggered on each API Records list request.
Could be used to validate or modify the response before returning it to the client.
Note that if you want to hide existing or add new computed Record fields prefer using the
`OnRecordEnrich`
hook because it is less error-prone and it is triggered
by all builtin Record responses (including when sending realtime Record events).
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every collection
app.OnRecordsListRequest().BindFunc(func(e *core.RecordsListRequestEvent) error {
// e.App
// e.Collection
// e.Records
// e.Result
// and all RequestEvent fields...
// fires only for "users" and "articles" collections
app.OnRecordsListRequest("users", "articles").BindFunc(func(e *core.RecordsListRequestEvent) error {
// e.App
// e.Collection
// e.Records
// e.Result
// and all RequestEvent fields...
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnRecordViewRequest
OnRecordViewRequest hook is triggered on each API Record view request.
Could be used to validate or modify the response before returning it to the client.
Note that if you want to hide existing or add new computed Record fields prefer using the
`OnRecordEnrich`
hook because it is less error-prone and it is triggered
by all builtin Record responses (including when sending realtime Record events).
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every collection
app.OnRecordViewRequest().BindFunc(func(e *core.RecordRequestEvent) error {
// e.App
// e.Collection
// e.Record
// and all RequestEvent fields...
// fires only for "users" and "articles" collections
app.OnRecordViewRequest("users", "articles").BindFunc(func(e *core.RecordRequestEvent) error {
log.Println(e.HttpContext)
log.Println(e.Record)
return nil
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnRecordCreateRequest
OnRecordCreateRequest hook is triggered on each API Record create request.
Could be used to additionally validate the request data or implement
completely different persistence behavior.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every collection
app.OnRecordCreateRequest().BindFunc(func(e *core.RecordRequestEvent) error {
// e.App
// e.Collection
// e.Record
// and all RequestEvent fields...
// fires only for "users" and "articles" collections
app.OnRecordCreateRequest("users", "articles").BindFunc(func(e *core.RecordRequestEvent) error {
// e.App
// e.Collection
// e.Record
// and all RequestEvent fields...
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnRecordUpdateRequest
OnRecordUpdateRequest hook is triggered on each API Record update request.
Could be used to additionally validate the request data or implement
completely different persistence behavior.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every collection
app.OnRecordUpdateRequest().BindFunc(func(e *core.RecordRequestEvent) error {
// e.App
// e.Collection
// e.Record
// and all RequestEvent fields...
// fires only for "users" and "articles" collections
app.OnRecordUpdateRequest("users", "articles").BindFunc(func(e *core.RecordRequestEvent) error {
// e.App
// e.Collection
// e.Record
// and all RequestEvent fields...
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnRecordDeleteRequest
OnRecordDeleteRequest hook is triggered on each API Record delete request.
Could be used to additionally validate the request data or implement
completely different delete behavior.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every collection
app.OnRecordDeleteRequest().BindFunc(func(e *core.RecordRequestEvent) error {
// e.App
// e.Collection
// e.Record
// and all RequestEvent fields...
// fires only for "users" and "articles" collections
app.OnRecordDeleteRequest("users", "articles").BindFunc(func(e *core.RecordRequestEvent) error {
// e.App
// e.Collection
// e.Record
// and all RequestEvent fields...
if err := app.Start(); err != nil {
log.Fatal(err)
}   ###### Record auth request hooks
OnRecordAuthRequest
OnRecordAuthRequest hook is triggered on each successful API
record authentication request (sign-in, token refresh, etc.).
Could be used to additionally validate or modify the authenticated
record data and token.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every auth collection
app.OnRecordAuthRequest().BindFunc(func(e *core.RecordAuthRequestEvent) error {
// e.App
// e.Record
// e.Token
// e.Meta
// e.AuthMethod
// and all RequestEvent fields...
// fires only for "users" and "managers" auth collections
app.OnRecordAuthRequest("users", "managers").BindFunc(func(e *core.RecordAuthRequestEvent) error {
// e.App
// e.Record
// e.Token
// e.Meta
// e.AuthMethod
// and all RequestEvent fields...
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnRecordAuthRefreshRequest
OnRecordAuthRefreshRequest hook is triggered on each Record
auth refresh API request (right before generating a new auth token).
Could be used to additionally validate the request data or implement
completely different auth refresh behavior.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every auth collection
app.OnRecordAuthRefreshRequest().BindFunc(func(e *core.RecordAuthWithOAuth2RequestEvent) error {
// e.App
// e.Collection
// e.Record
// and all RequestEvent fields...
// fires only for "users" and "managers" auth collections
app.OnRecordAuthRefreshRequest("users", "managers").BindFunc(func(e *core.RecordAuthWithOAuth2RequestEvent) error {
// e.App
// e.Collection
// e.Record
// and all RequestEvent fields...
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnRecordAuthWithPasswordRequest
OnRecordAuthWithPasswordRequest hook is triggered on each
Record auth with password API request.
e.Record could be nil if no matching identity is found, allowing
you to manually locate a different Record model (by reassigning e.Record).
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every auth collection
app.OnRecordAuthWithPasswordRequest().BindFunc(func(e *core.RecordAuthWithPasswordRequestEvent) error {
// e.App
// e.Collection
// e.Record (could be nil)
// e.Identity
// e.IdentityField
// e.Password
// and all RequestEvent fields...
// fires only for "users" and "managers" auth collections
app.OnRecordAuthWithPasswordRequest("users", "managers").BindFunc(func(e *core.RecordAuthWithPasswordRequestEvent) error {
// e.App
// e.Collection
// e.Record (could be nil)
// e.Identity
// e.IdentityField
// e.Password
// and all RequestEvent fields...
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnRecordAuthWithOAuth2Request
OnRecordAuthWithOAuth2Request hook is triggered on each Record
OAuth2 sign-in/sign-up API request (after token exchange and before external provider linking).
If e.Record is not set, then the OAuth2
request will try to create a new auth record.
To assign or link a different existing record model you can
change the e.Record field.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every auth collection
app.OnRecordAuthWithOAuth2Request().BindFunc(func(e *core.RecordAuthWithOAuth2RequestEvent) error {
// e.App
// e.Collection
// e.ProviderName
// e.ProviderClient
// e.Record (could be nil)
// e.OAuth2User
// e.CreateData
// e.IsNewRecord
// and all RequestEvent fields...
// fires only for "users" and "managers" auth collections
app.OnRecordAuthWithOAuth2Request("users", "managers").BindFunc(func(e *core.RecordAuthWithOAuth2RequestEvent) error {
// e.App
// e.Collection
// e.ProviderName
// e.ProviderClient
// e.Record (could be nil)
// e.OAuth2User
// e.CreateData
// e.IsNewRecord
// and all RequestEvent fields...
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnRecordRequestPasswordResetRequest
OnRecordRequestPasswordResetRequest hook is triggered on
each Record request password reset API request.
Could be used to additionally validate the request data or implement
completely different password reset behavior.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every auth collection
app.OnRecordRequestPasswordResetRequest().BindFunc(func(e *core.RecordRequestPasswordResetRequestEvent) error {
// e.App
// e.Collection
// e.Record
// and all RequestEvent fields...
// fires only for "users" and "managers" auth collections
app.OnRecordRequestPasswordResetRequest("users", "managers").BindFunc(func(e *core.RecordRequestPasswordResetRequestEvent) error {
// e.App
// e.Collection
// e.Record
// and all RequestEvent fields...
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnRecordConfirmPasswordResetRequest
OnRecordConfirmPasswordResetRequest hook is triggered on
each Record confirm password reset API request.
Could be used to additionally validate the request data or implement
completely different persistence behavior.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every auth collection
app.OnRecordConfirmPasswordResetRequest().BindFunc(func(e *core.RecordConfirmPasswordResetRequestEvent) error {
// e.App
// e.Collection
// e.Record
// and all RequestEvent fields...
// fires only for "users" and "managers" auth collections
app.OnRecordConfirmPasswordResetRequest("users", "managers").BindFunc(func(e *core.RecordConfirmPasswordResetRequestEvent) error {
// e.App
// e.Collection
// e.Record
// and all RequestEvent fields...
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnRecordRequestVerificationRequest
OnRecordRequestVerificationRequest hook is triggered on
each Record request verification API request.
Could be used to additionally validate the loaded request data or implement
completely different verification behavior.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every auth collection
app.OnRecordRequestVerificationRequest().BindFunc(func(e *core.RecordRequestVerificationRequestEvent) error {
// e.App
// e.Collection
// e.Record
// and all RequestEvent fields...
// fires only for "users" and "managers" auth collections
app.OnRecordRequestVerificationRequest("users", "managers").BindFunc(func(e *core.RecordRequestVerificationRequestEvent) error {
// e.App
// e.Collection
// e.Record
// and all RequestEvent fields...
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnRecordConfirmVerificationRequest
OnRecordConfirmVerificationRequest hook is triggered on each
Record confirm verification API request.
Could be used to additionally validate the request data or implement
completely different persistence behavior.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every auth collection
app.OnRecordConfirmVerificationRequest().BindFunc(func(e *core.RecordConfirmVerificationRequestEvent) error {
// e.App
// e.Collection
// e.Record
// and all RequestEvent fields...
// fires only for "users" and "managers" auth collections
app.OnRecordConfirmVerificationRequest("users", "managers").BindFunc(func(e *core.RecordConfirmVerificationRequestEvent) error {
// e.App
// e.Collection
// e.Record
// and all RequestEvent fields...
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnRecordRequestEmailChangeRequest
OnRecordRequestEmailChangeRequest hook is triggered on each
Record request email change API request.
Could be used to additionally validate the request data or implement
completely different request email change behavior.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every auth collection
app.OnRecordRequestEmailChangeRequest().BindFunc(func(e *core.RecordRequestEmailChangeRequestEvent) error {
// e.App
// e.Collection
// e.Record
// e.NewEmail
// and all RequestEvent fields...
// fires only for "users" and "managers" auth collections
app.OnRecordRequestEmailChangeRequest("users", "managers").BindFunc(func(e *core.RecordRequestEmailChangeRequestEvent) error {
// e.App
// e.Collection
// e.Record
// e.NewEmail
// and all RequestEvent fields...
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnRecordConfirmEmailChangeRequest
OnRecordConfirmEmailChangeRequest hook is triggered on each
Record confirm email change API request.
Could be used to additionally validate the request data or implement
completely different persistence behavior.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every auth collection
app.OnRecordConfirmEmailChangeRequest().BindFunc(func(e *core.RecordConfirmEmailChangeRequestEvent) error {
// e.App
// e.Collection
// e.Record
// e.NewEmail
// and all RequestEvent fields...
// fires only for "users" and "managers" auth collections
app.OnRecordConfirmEmailChangeRequest("users", "managers").BindFunc(func(e *core.RecordConfirmEmailChangeRequestEvent) error {
// e.App
// e.Collection
// e.Record
// e.NewEmail
// and all RequestEvent fields...
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnRecordRequestOTPRequest
OnRecordRequestOTPRequest hook is triggered on each Record
request OTP API request.
e.Record could be nil if no user with the requested email is found, allowing
you to manually create a new Record or locate a different Record model (by reassigning e.Record).
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every auth collection
app.OnRecordRequestOTPRequest().BindFunc(func(e *core.RecordCreateOTPRequestEvent) error {
// e.App
// e.Collection
// e.Record (could be nil)
// e.Password
// and all RequestEvent fields...
// fires only for "users" and "managers" auth collections
app.OnRecordRequestOTPRequest("users", "managers").BindFunc(func(e *core.RecordCreateOTPRequestEvent) error {
// e.App
// e.Collection
// e.Record (could be nil)
// e.Password
// and all RequestEvent fields...
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnRecordAuthWithOTPRequest
OnRecordAuthWithOTPRequest hook is triggered on each Record
auth with OTP API request.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every auth collection
app.OnRecordAuthWithOTPRequest().BindFunc(func(e *core.RecordAuthWithOTPRequestEvent) error {
// e.App
// e.Collection
// e.Record
// e.OTP
// and all RequestEvent fields...
// fires only for "users" and "managers" auth collections
app.OnRecordAuthWithOTPRequest("users", "managers").BindFunc(func(e *core.RecordAuthWithOTPRequestEvent) error {
// e.App
// e.Collection
// e.Record
// e.OTP
// and all RequestEvent fields...
if err := app.Start(); err != nil {
log.Fatal(err)
}   ###### Batch request hooks
OnBatchRequest
OnBatchRequest hook is triggered on each API batch request.
Could be used to additionally validate or modify the submitted batch requests.
This hook will also fire the corresponding OnRecordCreateRequest, OnRecordUpdateRequest, OnRecordDeleteRequest hooks, where e.App is the batch transactional app.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
app.OnBatchRequest().BindFunc(func(e *core.BatchRequestEvent) error {
// e.App
// e.Batch
// and all RequestEvent fields...
if err := app.Start(); err != nil {
log.Fatal(err)
}   ###### File request hooks
Could be used to validate or modify the file response before returning it to the client.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// e.App
// e.Collection
// e.Record
// e.FileField
// e.ServedPath
// e.ServedName
// and all RequestEvent fields...
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnFileTokenRequest
OnFileTokenRequest hook is triggered on each auth file token API request.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every auth model
app.OnFileTokenRequest().BindFunc(func(e *core.FileTokenRequestEvent) error {
// e.App
// e.Record
// e.Token
// and all RequestEvent fields...
// fires only for "users"
app.OnFileTokenRequest("users").BindFunc(func(e *core.FileTokenRequestEvent) error {
// e.App
// e.Record
// e.Token
// and all RequestEvent fields...
if err := app.Start(); err != nil {
log.Fatal(err)
}   ###### Collection request hooks
OnCollectionsListRequest
`OnCollectionsListRequest` hook is triggered on each API Collections list request.
Could be used to validate or modify the response before returning it to the client.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
app.OnCollectionsListRequest().BindFunc(func(e *core.CollectionsListRequestEvent) error {
// e.App
// e.Collections
// e.Result
// and all RequestEvent fields...
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnCollectionViewRequest
`OnCollectionViewRequest` hook is triggered on each API Collection view request.
Could be used to validate or modify the response before returning it to the client.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
app.OnCollectionViewRequest().BindFunc(func(e *core.CollectionRequestEvent) error {
// e.App
// e.Collection
// and all RequestEvent fields...
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnCollectionCreateRequest
`OnCollectionCreateRequest` hook is triggered on each API Collection create request.
Could be used to additionally validate the request data or implement
completely different persistence behavior.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
app.OnCollectionCreateRequest().BindFunc(func(e *core.CollectionRequestEvent) error {
// e.App
// e.Collection
// and all RequestEvent fields...
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnCollectionUpdateRequest
`OnCollectionUpdateRequest` hook is triggered on each API Collection update request.
Could be used to additionally validate the request data or implement
completely different persistence behavior.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
app.OnCollectionUpdateRequest().BindFunc(func(e *core.CollectionRequestEvent) error {
// e.App
// e.Collection
// and all RequestEvent fields...
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnCollectionDeleteRequest
`OnCollectionDeleteRequest` hook is triggered on each API Collection delete request.
Could be used to additionally validate the request data or implement
completely different delete behavior.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
app.OnCollectionDeleteRequest().BindFunc(func(e *core.CollectionRequestEvent) error {
// e.App
// e.Collection
// and all RequestEvent fields...
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnCollectionsImportRequest
`OnCollectionsImportRequest` hook is triggered on each API
collections import request.
Could be used to additionally validate the imported collections or
to implement completely different import behavior.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
app.OnCollectionsImportRequest().BindFunc(func(e *core.CollectionsImportRequestEvent) error {
// e.App
// e.CollectionsData
// e.DeleteMissing
if err := app.Start(); err != nil {
log.Fatal(err)
}   ###### Settings request hooks
OnSettingsListRequest
`OnSettingsListRequest` hook is triggered on each API Settings list request.
Could be used to validate or modify the response before returning it to the client.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
app.OnSettingsListRequest().BindFunc(func(e *core.SettingsListRequestEvent) error {
// e.App
// e.Settings
// and all RequestEvent fields...
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnSettingsUpdateRequest
`OnSettingsUpdateRequest` hook is triggered on each API Settings update request.
Could be used to additionally validate the request data or
implement completely different persistence behavior.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
app.OnSettingsUpdateRequest().BindFunc(func(e *core.SettingsUpdateRequestEvent) error {
// e.App
// e.OldSettings
// e.NewSettings
// and all RequestEvent fields...
if err := app.Start(); err != nil {
log.Fatal(err)
}   ### Base model hooks
The Model hooks are fired for all PocketBase structs that implements the Model DB interface - Record, Collection, Log, etc.
For convenience, if you want to listen to only the Record or Collection DB model
events without doing manual type assertion, you can use the
OnRecord*
and
OnCollection*
proxy hooks above.
OnModelValidate
OnModelValidate is called every time when a Model is being validated,
e.g. triggered by App.Validate() or App.Save().
For convenience, if you want to listen to only the Record or Collection models
events without doing manual type assertion, you can use the equivalent OnRecord* and OnCollection* proxy hooks.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every model
app.OnModelValidate().BindFunc(func(e *core.ModelEvent) error {
// e.App
// e.Model
// fires only for "users" and "articles" models
app.OnModelValidate("users", "articles").BindFunc(func(e *core.ModelEvent) error {
// e.App
// e.Model
if err := app.Start(); err != nil {
log.Fatal(err)
}   ###### Base model create hooks
OnModelCreate
OnModelCreate is triggered every time when a new Model is being created,
e.g. triggered by App.Save().
and the INSERT DB statement.
and the INSERT DB statement.
Note that successful execution doesn't guarantee that the Model
is persisted in the database since its wrapping transaction may
not have been committed yet.
If you want to listen to only the actual persisted events, you can
bind to `OnModelAfterCreateSuccess` or `OnModelAfterCreateError` hooks.
For convenience, if you want to listen to only the Record or Collection models
events without doing manual type assertion, you can use the equivalent OnRecord* and OnCollection* proxy hooks.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every model
app.OnModelCreate().BindFunc(func(e *core.ModelEvent) error {
// e.App
// e.Model
// fires only for "users" and "articles" models
app.OnModelCreate("users", "articles").BindFunc(func(e *core.ModelEvent) error {
// e.App
// e.Model
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnModelCreateExecute
OnModelCreateExecute is triggered after successful Model validation
and right before the model INSERT DB statement execution.
Usually it is triggered as part of the App.Save() in the following firing order:
OnModelCreate
&nbsp;->
OnModelValidate (skipped with App.SaveNoValidate())
&nbsp;->
OnModelCreateExecute
Note that successful execution doesn't guarantee that the Model
is persisted in the database since its wrapping transaction may
not have been committed yet.
If you want to listen to only the actual persisted events, you can
bind to `OnModelAfterCreateSuccess` or `OnModelAfterCreateError` hooks.
For convenience, if you want to listen to only the Record or Collection models
events without doing manual type assertion, you can use the equivalent OnRecord* and OnCollection* proxy hooks.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every model
app.OnModelCreateExecute().BindFunc(func(e *core.ModelEvent) error {
// e.App
// e.Model
// fires only for "users" and "articles" models
app.OnModelCreateExecute("users", "articles").BindFunc(func(e *core.ModelEvent) error {
// e.App
// e.Model
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnModelAfterCreateSuccess
OnModelAfterCreateSuccess is triggered after each successful
Model DB create persistence.
Note that when a Model is persisted as part of a transaction,
this hook is delayed and executed only AFTER the transaction has been committed.
This hook is NOT triggered in case the transaction fails/rollbacks.
For convenience, if you want to listen to only the Record or Collection models
events without doing manual type assertion, you can use the equivalent OnRecord* and OnCollection* proxy hooks.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every model
app.OnModelAfterCreateSuccess().BindFunc(func(e *core.ModelEvent) error {
// e.App
// e.Model
// fires only for "users" and "articles" models
app.OnModelAfterCreateSuccess("users", "articles").BindFunc(func(e *core.ModelEvent) error {
// e.App
// e.Model
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnModelAfterCreateError
OnModelAfterCreateError is triggered after each failed
Model DB create persistence.
Note that the execution of this hook is either immediate or delayed
depending on the error:
-immediate on App.Save() failure
-delayed on transaction rollback
For convenience, if you want to listen to only the Record or Collection models
events without doing manual type assertion, you can use the equivalent OnRecord* and OnCollection* proxy hooks.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every model
app.OnModelAfterCreateError().BindFunc(func(e *core.ModelErrorEvent) error {
// e.App
// e.Model
// e.Error
// fires only for "users" and "articles" models
app.OnModelAfterCreateError("users", "articles").BindFunc(func(e *core.ModelErrorEvent) error {
// e.App
// e.Model
// e.Error
if err := app.Start(); err != nil {
log.Fatal(err)
}   ###### Base model update hooks
OnModelUpdate
OnModelUpdate is triggered every time when a new Model is being updated,
e.g. triggered by App.Save().
and the UPDATE DB statement.
and the UPDATE DB statement.
Note that successful execution doesn't guarantee that the Model
is persisted in the database since its wrapping transaction may
not have been committed yet.
If you want to listen to only the actual persisted events, you can
bind to `OnModelAfterUpdateSuccess` or `OnModelAfterUpdateError` hooks.
For convenience, if you want to listen to only the Record or Collection models
events without doing manual type assertion, you can use the equivalent OnRecord* and OnCollection* proxy hooks.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every model
app.OnModelUpdate().BindFunc(func(e *core.ModelEvent) error {
// e.App
// e.Model
// fires only for "users" and "articles" models
app.OnModelUpdate("users", "articles").BindFunc(func(e *core.ModelEvent) error {
// e.App
// e.Model
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnModelUpdateExecute
OnModelUpdateExecute is triggered after successful Model validation
and right before the model UPDATE DB statement execution.
Usually it is triggered as part of the App.Save() in the following firing order:
OnModelUpdate
&nbsp;->
OnModelValidate (skipped with App.SaveNoValidate())
&nbsp;->
OnModelUpdateExecute
Note that successful execution doesn't guarantee that the Model
is persisted in the database since its wrapping transaction may
not have been committed yet.
If you want to listen to only the actual persisted events, you can
bind to `OnModelAfterUpdateSuccess` or `OnModelAfterUpdateError` hooks.
For convenience, if you want to listen to only the Record or Collection models
events without doing manual type assertion, you can use the equivalent OnRecord* and OnCollection* proxy hooks.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every model
app.OnModelUpdateExecute().BindFunc(func(e *core.ModelEvent) error {
// e.App
// e.Model
// fires only for "users" and "articles" models
app.OnModelUpdateExecute("users", "articles").BindFunc(func(e *core.ModelEvent) error {
// e.App
// e.Model
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnModelAfterUpdateSuccess
OnModelAfterUpdateSuccess is triggered after each successful
Model DB update persistence.
Note that when a Model is persisted as part of a transaction,
this hook is delayed and executed only AFTER the transaction has been committed.
This hook is NOT triggered in case the transaction fails/rollbacks.
For convenience, if you want to listen to only the Record or Collection models
events without doing manual type assertion, you can use the equivalent OnRecord* and OnCollection* proxy hooks.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every model
app.OnModelAfterUpdateSuccess().BindFunc(func(e *core.ModelEvent) error {
// e.App
// e.Model
// fires only for "users" and "articles" models
app.OnModelAfterUpdateSuccess("users", "articles").BindFunc(func(e *core.ModelEvent) error {
// e.App
// e.Model
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnModelAfterUpdateError
OnModelAfterUpdateError is triggered after each failed
Model DB update persistence.
Note that the execution of this hook is either immediate or delayed
depending on the error:
-immediate on App.Save() failure
-delayed on transaction rollback
For convenience, if you want to listen to only the Record or Collection models
events without doing manual type assertion, you can use the equivalent OnRecord* and OnCollection* proxy hooks.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every model
app.OnModelAfterUpdateError().BindFunc(func(e *core.ModelErrorEvent) error {
// e.App
// e.Model
// e.Error
// fires only for "users" and "articles" models
app.OnModelAfterUpdateError("users", "articles").BindFunc(func(e *core.ModelErrorEvent) error {
// e.App
// e.Model
// e.Error
if err := app.Start(); err != nil {
log.Fatal(err)
}   ###### Base model delete hooks
OnModelDelete
OnModelDelete is triggered every time when a new Model is being deleted,
e.g. triggered by App.Delete().
and the UPDATE DB statement.
and the UPDATE DB statement.
Note that successful execution doesn't guarantee that the Model
is deleted from the database since its wrapping transaction may
not have been committed yet.
If you want to listen to only the actual persisted deleted events, you can
bind to `OnModelAfterDeleteSuccess` or `OnModelAfterDeleteError` hooks.
For convenience, if you want to listen to only the Record or Collection models
events without doing manual type assertion, you can use the equivalent OnRecord* and OnCollection* proxy hooks.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every model
app.OnModelDelete().BindFunc(func(e *core.ModelEvent) error {
// e.App
// e.Model
// fires only for "users" and "articles" models
app.OnModelDelete("users", "articles").BindFunc(func(e *core.ModelEvent) error {
// e.App
// e.Model
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnModelDeleteExecute
OnModelDeleteExecute is triggered after the internal delete checks and
right before the Model the model DELETE DB statement execution.
Usually it is triggered as part of the App.Delete() in the following firing order:
OnModelDelete
&nbsp;->
internal delete checks
&nbsp;->
OnModelDeleteExecute
Note that successful execution doesn't guarantee that the Model
is deleted from the database since its wrapping transaction may
not have been committed yet.
If you want to listen to only the actual persisted events, you can
bind to `OnModelAfterDeleteSuccess` or `OnModelAfterDeleteError` hooks.
For convenience, if you want to listen to only the Record or Collection models
events without doing manual type assertion, you can use the equivalent OnRecord* and OnCollection* proxy hooks.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every model
app.OnModelDeleteExecute().BindFunc(func(e *core.ModelEvent) error {
// e.App
// e.Model
// fires only for "users" and "articles" models
app.OnModelDeleteExecute("users", "articles").BindFunc(func(e *core.ModelEvent) error {
// e.App
// e.Model
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnModelAfterDeleteSuccess
OnModelAfterDeleteSuccess is triggered after each successful
Model DB delete persistence.
Note that when a Model is deleted as part of a transaction,
this hook is delayed and executed only AFTER the transaction has been committed.
This hook is NOT triggered in case the transaction fails/rollbacks.
For convenience, if you want to listen to only the Record or Collection models
events without doing manual type assertion, you can use the equivalent OnRecord* and OnCollection* proxy hooks.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every model
app.OnModelAfterDeleteSuccess().BindFunc(func(e *core.ModelEvent) error {
// e.App
// e.Model
// fires only for "users" and "articles" models
app.OnModelAfterDeleteSuccess("users", "articles").BindFunc(func(e *core.ModelEvent) error {
// e.App
// e.Model
if err := app.Start(); err != nil {
log.Fatal(err)
}     OnModelAfterDeleteError
OnModelAfterDeleteError is triggered after each failed
Model DB delete persistence.
Note that the execution of this hook is either immediate or delayed
depending on the error:
-immediate on App.Delete() failure
-delayed on transaction rollback
For convenience, if you want to listen to only the Record or Collection models
events without doing manual type assertion, you can use the equivalent OnRecord* and OnCollection* proxy hooks.
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
// fires for every model
app.OnModelAfterDeleteError().BindFunc(func(e *core.ModelErrorEvent) error {
// e.App
// e.Model
// e.Error
// fires only for "users" and "articles" models
app.OnModelAfterDeleteError("users", "articles").BindFunc(func(e *core.ModelErrorEvent) error {
// e.App
// e.Model
// e.Error
if err := app.Start(); err != nil {
log.Fatal(err)

## 27.Extend with Go - Routing
GET /hello/{name}"
# Extend with Go - Routing
`GET /hello/{name}"`
Routing PocketBase routing is built on top of the standard Go
net/http.ServeMux.
The router can be accessed via the app.OnServe() hook allowing you to register custom endpoints
and middlewares.
### Routes
##### Registering new routes
Every route has a path, handler function and eventually middlewares attached to it. For example:
app.OnServe().BindFunc(func(se *core.ServeEvent) error {
// register "GET /hello/{name}" route (allowed for everyone)
se.Router.GET("/hello/{name}", func(e *core.RequestEvent) error {
name := e.Request.PathValue("name")
return e.String(http.StatusOK, "Hello " + name)
// register "POST /api/myapp/settings" route (allowed only for authenticated users)
se.Router.POST("/api/myapp/settings", func(e *core.RequestEvent) error {
// do something ...
return e.JSON(http.StatusOK, map[string]bool{"success": true})
}).Bind(apis.RequireAuth())
})  There are several routes registration methods available, but the most common ones are:
se.Router.GET(path, action)
se.Router.POST(path, action)
se.Router.PUT(path, action)
se.Router.PATCH(path, action)
se.Router.DELETE(path, action)
// If you want to handle any HTTP method define only a path (e.g. "/example")
// OR if you want to specify a custom one add it as prefix to the path (e.g. "TRACE /example")
se.Router.Any(pattern, action)  The router also supports creating groups for routes that share the same base path and middlewares. For
example:
g := se.Router.Group("/api/myapp")
// group middleware
g.Bind(apis.RequireAuth())
// group routes
g.GET("", action1)
g.GET("/example/{id}", action2)
g.PATCH("/example/{id}", action3).BindFunc(
/* custom route specific middleware func */
// nested group
sub := g.Group("/sub")
sub.GET("/sub1", action4)  The example registers the following endpoints
(all require authenticated user access):
-GET /api/myapp -&gt; action1
-GET /api/myapp/example/{id} -&gt; action2
-PATCH /api/myapp/example/{id} -&gt; action3
-GET /api/myapp/example/sub/sub1 -&gt; action4
Each router group and route could define middlewares in a similar manner to the
regular app hooks via the Bind/BindFunc methods, allowing you to perform various BEFORE or AFTER
action operations (e.g. inspecting request headers, custom access checks, etc.).
##### Path parameters and matching rules
Because PocketBase routing is based on top of the Go standard router mux, we follow the same pattern
matching rules. Below you could find a short overview but for more details please refer to
net/http.ServeMux.
In general, a route pattern looks like [METHOD ][HOST]/[PATH]
(the METHOD prefix is added automatically when using the designated GET(),
POST(), etc. methods)).
Route paths can include parameters in the format {paramName}.
You can also use {paramName...} format to specify a parameter that targets more than one path
segment.
A pattern ending with a trailing slash / acts as anonymous wildcard and matches any requests
that begins with the defined route. If you want to have a trailing slash but to indicate the end of the
URL then you need to end the path with the special
{$} parameter.
If your route path starts with /api/
consider combining it with your unique app name like /api/myapp/... to avoid collisions
with system routes.
Here are some examples:
// match "GET /index.html" (for any host)
se.Router.GET("/index.html")
// match "GET /static/", "GET /static/a/b/c", etc.
se.Router.GET("/static/")
// match "GET /static/", "GET /static/a/b/c", etc.
// (similar to the above but with a named wildcard parameter)
se.Router.GET("/static/{path...}")
// match only "GET /static/" (if no "/static" is registered, it is 301 redirected)
se.Router.GET("/static/{$}")
// match "GET /customers/john", "GET /customers/jane", etc.
se.Router.GET("/customers/{name}")   In the following examples e is usually
*core.RequestEvent value.
##### Reading path parameters
`id := e.Request.PathValue("id")`  ##### Retrieving the current auth state
The request auth state can be accessed (or set) via the RequestEvent.Auth field.
authRecord := e.Auth
isGuest := e.Auth == nil
// the same as "e.Auth != nil &amp;&amp; e.Auth.IsSuperuser()"
isSuperuser := e.HasSuperuserAuth()  Alternatively you could also access the request data from the summarized request info instance
(usually used in hooks like the OnRecordEnrich where there is no direct access to the request)
info, err := e.RequestInfo()
authRecord := info.Auth
isGuest := info.Auth == nil
// the same as "info.Auth != nil &amp;&amp; info.Auth.IsSuperuser()"
isSuperuser := info.HasSuperuserAuth()  ##### Reading query parameters
search := e.Request.URL.Query().Get("search")
// or via the parsed request info
info, err := e.RequestInfo()
search := info.Query["search"]  ##### Reading request headers
token := e.Request.Header.Get("Some-Header")
// or via the parsed request info
// (the header value is always normalized per the @request.headers.* API rules format)
info, err := e.RequestInfo()
token := info.Headers["some_header"]  ##### Writing response headers
`e.Response.Header().Set("Some-Header", "123")`  ##### Retrieving uploaded files
// retrieve the uploaded files and parse the found multipart data into a ready-to-use []*filesystem.File
files, err := e.FindUploadedFiles("document")
mf, mh, err := e.Request.FormFile("document")  ##### Reading request body
Body parameters can be read either via
e.BindBody
OR through the parsed request info (requires manual type assertions).
The e.BindBody argument must be a pointer to a struct or map[string]any.
The following struct tags are supported
(the specific binding rules and which one will be used depend on the request Content-Type):
-json (json body)- uses the builtin Go JSON package for unmarshaling.
-xml (xml body) - uses the builtin Go XML package for unmarshaling.
-form (form data) - utilizes the custom
router.UnmarshalRequestData
method.
NB! When binding structs make sure that they don&#39;t have public fields that shouldn&#39;t be bindable and it is
advisable such fields to be unexported or define a separate struct with just the safe bindable fields.
// read/scan the request body fields into a typed struct
data := struct {
// unexported to prevent binding
somethingPrivate string
Title       string `json:"title" form:"title"`
Description string `json:"description" form:"description"`
Active      bool   `json:"active" form:"active"`
if err := e.BindBody(&amp;data); err != nil {
return e.BadRequestError("Failed to read request data", err)
// alternatively, read the body via the parsed request info
info, err := e.RequestInfo()
title, ok := info.Body["title"].(string)  ##### Writing response body
For all supported methods, you can refer to
router.Event
// send response with JSON body
// (it also provides a generic response fields picker/filter if the "fields" query parameter is set)
e.JSON(http.StatusOK, map[string]any{"name": "John"})
// send response with string body
// send response with HTML body
// (check also the "Rendering templates" section)
e.HTML(http.StatusOK, "&lt;h1>Hello!&lt;/h1>")
// redirect
// send response with no body
e.NoContent(http.StatusNoContent)
// serve a single file
e.FileFS(os.DirFS("..."), "example.txt")
// stream the specified reader
e.Stream(http.StatusOK, "application/octet-stream", reader)
// send response with blob (bytes slice) body
e.Blob(http.StatusOK, "application/octet-stream", []byte{ ... })  ##### Reading the client IP
// The IP of the last client connecting to your server.
// The returned IP is safe and can be always trusted.
// When behind a reverse proxy (e.g. nginx) this method returns the IP of the proxy.
// https://pkg.go.dev/github.com/pocketbase/pocketbase/tools/router#Event.RemoteIP
ip := e.RemoteIP()
// The "real" IP of the client based on the configured Settings.TrustedProxy header(s).
// If such headers are not set, it fallbacks to e.RemoteIP().
// https://pkg.go.dev/github.com/pocketbase/pocketbase/core#RequestEvent.RealIP
ip := e.RealIP()  ##### Request store
The core.RequestEvent comes with a local store that you can use to share custom data between
middlewares and the route action.
// store for the duration of the request
e.Set("someKey", 123)
// retrieve later
val := e.Get("someKey").(int) // 123  ### Middlewares
##### Registering middlewares
Middlewares allow inspecting, intercepting and filtering route requests.
All middleware functions share the same signature with the route actions (aka.
want to proceed with the execution chain.
Middlewares can be registered globally, on group and on route level using the
Bind
and BindFunc methods.
Here is a minimal example of what a global middleware looks like:
app.OnServe().BindFunc(func(se *core.ServeEvent) error {
// register a global middleware
se.Router.BindFunc(func(e *core.RequestEvent) error {
if e.Request.Header.Get("Something") == "" {
return e.BadRequestError("Something header value is missing!", nil)
})  RouterGroup.Bind(middlewares...)
Route.Bind(middlewares...)
registers one or more middleware handlers.
Similar to the other app hooks, a middleware handler has 3 fields:
-Id (optional) - the name of the middleware (could be used as argument for
Unbind)
-Priority (optional) - the execution order of the middleware (if empty fallbacks to
the order of registration in the code)
-Func (required) - the middleware handler function
Often you don&#39;t need to specify the Id or Priority of the middleware and for
convenience you can instead use directly
RouterGroup.BindFunc(funcs...)
Route.BindFunc(funcs...) .
Below is a slightly more advanced example showing all options and the execution sequence (2,0,1,3,4):
app.OnServe().BindFunc(func(se *core.ServeEvent) error {
// attach global middleware
se.Router.BindFunc(func(e *core.RequestEvent) error {
println(0)
g := se.Router.Group("/sub")
// attach group middleware
g.BindFunc(func(e *core.RequestEvent) error {
println(1)
// attach group middleware with an id and custom priority
g.Bind(&amp;hook.Handler[*core.RequestEvent]{
Id: "something",
Func: func(e *core.RequestEvent) error {
println(2)
Priority: -1,
// attach middleware to a single route
// "GET /sub/hello" should print the sequence: 2,0,1,3,4
g.GET("/hello", func(e *core.RequestEvent) error {
println(4)
return e.String(200, "Hello!")
}).BindFunc(func(e *core.RequestEvent) error {
println(3)
})  ##### Removing middlewares
To remove a registered middleware from the execution chain for a specific group or route you can make use
of the
Unbind(id) method.
Note that only middlewares that have a non-empty Id can be removed.
app.OnServe().BindFunc(func(se *core.ServeEvent) error {
// global middleware
se.Router.Bind(&amp;hook.Handler[*core.RequestEvent]{
Id: "test",
Func: func(e *core.RequestEvent) error {
// "GET /A" invokes the "test" middleware
se.Router.GET("/A", func(e *core.RequestEvent) error {
return e.String(200, "A")
// "GET /B" doesn't invoke the "test" middleware
se.Router.GET("/B", func(e *core.RequestEvent) error {
return e.String(200, "B")
}).Unbind("test")
})  ##### Builtin middlewares
The
apis
package exposes several middlewares that you can use as part of your application.
// Require the request client to be unauthenticated (aka. guest).
// Example: Route.Bind(apis.RequireGuestOnly())
apis.RequireGuestOnly()
// Require the request client to be authenticated
// (optionally specify a list of allowed auth collection names, default to any).
// Example: Route.Bind(apis.RequireAuth())
apis.RequireAuth(optCollectionNames...)
// Require the request client to be authenticated as superuser
// (this is an alias for apis.RequireAuth(core.CollectionNameSuperusers)).
// Example: Route.Bind(apis.RequireSuperuserAuth())
apis.RequireSuperuserAuth()
// Require the request client to be authenticated as superuser OR
// regular auth record with id matching the specified route parameter (default to "id").
// Example: Route.Bind(apis.RequireSuperuserOrOwnerAuth(""))
apis.RequireSuperuserOrOwnerAuth(ownerIdParam)
// Changes the global 32MB default request body size limit (set it to 0 for no limit).
// Note that system record routes have dynamic body size limit based on their collection field types.
// Example: Route.Bind(apis.BodyLimit(10 &lt;&lt; 20))
apis.BodyLimit(limitBytes)
// Compresses the HTTP response using Gzip compression scheme.
// Example: Route.Bind(apis.Gzip())
apis.Gzip()
// Instructs the activity logger to log only requests that have failed/returned an error.
// Example: Route.Bind(apis.SkipSuccessActivityLog())
apis.SkipSuccessActivityLog()  ##### Default globally registered middlewares
The below list is mostly useful for users that may want to plug their own custom middlewares before/after
the priority of the default global ones, for example: registering a custom auth loader before the rate
limiter with `apis.DefaultRateLimitMiddlewarePriority - 1` so that the rate limit can be applied
properly based on the loaded auth state. All PocketBase applications have the below internal middlewares registered out of the box (sorted by their priority):
-WWW redirect apis.DefaultWWWRedirectMiddlewareId apis.DefaultWWWRedirectMiddlewarePriority  Performs www -&gt; non-www redirect(s) if the request host matches with one of the values in
certificate host policy.
-CORS apis.DefaultCorsMiddlewareId apis.DefaultCorsMiddlewarePriority  By default all origins are allowed (PocketBase is stateless and doesn&#39;t rely on cookies) and can
be configured with the --origins
flag but for more advanced customization it can be also replaced entirely by binding with
apis.CORS(config) middleware or registering your own custom one in its place.
-Activity logger apis.DefaultActivityLoggerMiddlewareId apis.DefaultActivityLoggerMiddlewarePriority  Saves request information into the logs auxiliary database.
-Auto panic recover apis.DefaultPanicRecoverMiddlewareId apis.DefaultPanicRecoverMiddlewarePriority  Default panic-recover handler.
-Auth token loader apis.DefaultLoadAuthTokenMiddlewareId apis.DefaultLoadAuthTokenMiddlewarePriority  Loads the auth token from the Authorization header and populates the related auth
record into the request event (aka. e.Auth).
-Security response headers apis.DefaultSecurityHeadersMiddlewareId apis.DefaultSecurityHeadersMiddlewarePriority  Adds default common security headers (X-XSS-Protection,
X-Content-Type-Options,
X-Frame-Options) to the response (can be overwritten by other middlewares or from
inside the route action).
-Rate limit apis.DefaultRateLimitMiddlewareId apis.DefaultRateLimitMiddlewarePriority  Rate limits client requests based on the configured app settings (it does nothing if the rate
limit option is not enabled).
-Body limit apis.DefaultBodyLimitMiddlewareId apis.DefaultBodyLimitMiddlewarePriority  Applies a default max ~32MB request body limit for all custom routes ( system record routes have
dynamic body size limit based on their collection field types). Can be overwritten on group/route
level by simply rebinding the apis.BodyLimit(limitBytes) middleware.
### Error response
PocketBase has a global error handler and every returned error from a route or middleware will be safely
converted by default to a generic ApiError to avoid accidentally leaking sensitive
in --dev mode).
To make it easier returning formatted JSON error responses, the request event provides several
ApiError methods.
router.SafeErrorItem/validation.Error items.
import validation "github.com/go-ozzo/ozzo-validation/v4"
se.Router.GET("/example", func(e *core.RequestEvent) error {
// construct ApiError with custom status code and validation data error
return e.Error(500, "something went wrong", map[string]validation.Error{
"title": validation.NewError("invalid_title", "Invalid or missing title"),
// if message is empty string, a default one will be set
return e.BadRequestError(optMessage, optData)      // 400 ApiError
return e.UnauthorizedError(optMessage, optData)    // 401 ApiError
return e.ForbiddenError(optMessage, optData)       // 403 ApiError
return e.NotFoundError(optMessage, optData)        // 404 ApiError
return e.TooManyRequestsError(optMessage, optData) // 429 ApiError
return e.InternalServerError(optMessage, optData)  // 500 ApiError
})  This is not very common but if you want to return ApiError outside of request related
handlers, you can use the below
apis.* factories:
import (
validation "github.com/go-ozzo/ozzo-validation/v4"
"github.com/pocketbase/pocketbase/apis"
app.OnRecordCreate().BindFunc(func(e *core.RecordEvent) error {
// construct ApiError with custom status code and validation data error
return apis.NewApiError(500, "something went wrong", map[string]validation.Error{
"title": validation.NewError("invalid_title", "Invalid or missing title"),
// if message is empty string, a default one will be set
return apis.NewBadRequestError(optMessage, optData)      // 400 ApiError
return apis.NewUnauthorizedError(optMessage, optData)    // 401 ApiError
return apis.NewForbiddenError(optMessage, optData)       // 403 ApiError
return apis.NewNotFoundError(optMessage, optData)        // 404 ApiError
return apis.NewTooManyRequestsError(optMessage, optData) // 429 ApiError
return apis.NewInternalServerError(optMessage, optData)  // 500 ApiError
})  ### Helpers
##### Serving static directory
apis.Static()
serves static directory content from fs.FS instance.
Expects the route to have a {path...} wildcard parameter.
app.OnServe().BindFunc(func(se *core.ServeEvent) error {
// serves static files from the provided dir (if exists)
se.Router.GET("/{path...}", apis.Static(os.DirFS("/path/to/public"), false))
})  ##### Auth response
apis.RecordAuthResponse()
writes standardized JSON record auth response (aka. token + record data) into the specified request body.
Could be used as a return result from a custom auth route.
app.OnServe().BindFunc(func(se *core.ServeEvent) error {
se.Router.POST("/phone-login", func(e *core.RequestEvent) error {
data := struct {
Phone    string `json:"phone" form:"phone"`
Password string `json:"password" form:"password"`
if err := e.BindBody(&amp;data); err != nil {
return e.BadRequestError("Failed to read request data", err)
record, err := e.App.FindFirstRecordByData("users", "phone", data.Phone)
if err != nil || !record.ValidatePassword(data.Password) {
// return generic 400 error as a basic enumeration protection
return e.BadRequestError("Invalid credentials", err)
return apis.RecordAuthResponse(e, record, "phone", nil)
})  ##### Enrich record(s)
apis.EnrichRecord()
and
apis.EnrichRecords()
helpers parses the request context and enrich the provided record(s) by:
-expands relations (if defaultExpands and/or ?expand query parameter is set)
-ensures that the emails of the auth record and its expanded auth relations are visible only for the
current logged superuser, record owner or record with manage access
app.OnServe().BindFunc(func(se *core.ServeEvent) error {
se.Router.GET("/custom-article", func(e *core.RequestEvent) error {
records, err := e.App.FindRecordsByFilter("article", "status = 'active'", "-created", 40, 0)
if err != nil {
return e.NotFoundError("No active articles", err)
// enrich the records with the "categories" relation as default expand
err = apis.EnrichRecords(e, records, "categories")
if err != nil {
return err
return e.JSON(http.StatusOK, records)
})  ##### Go http.Handler wrappers
If you want to register standard Go http.Handler function and middlewares, you can use
apis.WrapStdHandler(handler)
and
apis.WrapStdMiddleware(func)
functions.
### Sending request to custom routes using the SDKs
The official PocketBase SDKs expose the internal send() method that could be used to send requests
to your custom route(s).
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
await pb.send("/hello", {
// for other options check
// https://developer.mozilla.org/en-US/docs/Web/API/fetch#options
query: { "abc": 123 },
});  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
await pb.send("/hello", query: { "abc": 123 })

## 28.Extend with Go - Database
# Extend with Go - Database
Database core.App
is the main interface to interact with the database.
App.DB() returns a dbx.Builder that can run all kinds of SQL statements, including
Most of the common DB operations are listed below, but you can find further information in the
dbx package godoc
For more details and examples how to interact with Record and Collection models programmatically
you could also check Collection operations
and
Record operations sections.
### Executing queries
To execute DB queries you can start with the NewQuery(&quot;...&quot;) statement and then call one of:
-Execute()
- for any query statement that is not meant to retrieve data:
res, err := app.DB().
NewQuery("DELETE FROM articles WHERE status = 'archived'").
Execute()
-One()
- to populate a single row into a struct:
type User struct {
Id     string                  `db:"id" json:"id"`
Status bool                    `db:"status" json:"status"`
Age    int                     `db:"age" json:"age"`
Roles  types.JSONArray[string] `db:"roles" json:"roles"`
user := User{}
err := app.DB().
NewQuery("SELECT id, status, age, roles FROM users WHERE id=1").
One(&amp;user)
-All()
- to populate multiple rows into a slice of structs:
type User struct {
Id     string                  `db:"id" json:"id"`
Status bool                    `db:"status" json:"status"`
Age    int                     `db:"age" json:"age"`
Roles  types.JSONArray[string] `db:"roles" json:"roles"`
users := []User{}
err := app.DB().
NewQuery("SELECT id, status, age, roles FROM users LIMIT 100").
All(&amp;users)
### Binding parameters
To prevent SQL injection attacks, you should use named parameters for any expression value that comes from
user input. This could be done using the named {:paramName}
Bind(params). For example:
type Post struct {
Name     string         `db:"name" json:"name"`
Created  types.DateTime `db:"created" json:"created"`
posts := []Post{}
err := app.DB().
NewQuery("SELECT name, created FROM posts WHERE created >= {:from} and created &lt;= {:to}").
Bind(dbx.Params{
"from": "2023-06-25 00:00:00.000Z",
"to":   "2023-06-28 23:59:59.999Z",
All(&amp;posts)  ### Query builder
Instead of writing plain SQLs, you can also compose SQL statements programmatically using the db query
builder.
Every SQL keyword has a corresponding query building method. For example, SELECT corresponds
to Select(), FROM corresponds to From(),
WHERE corresponds to Where(), and so on.
users := []struct {
Id    string `db:"id" json:"id"`
Email string `db:"email" json:"email"`
app.DB().
Select("id", "email").
From("users").
Limit(100).
OrderBy("created ASC").
All(&amp;users)  ##### Select(), AndSelect(), Distinct()
The Select(...cols) method initializes a SELECT query builder. It accepts a list
of the column names to be selected.
To add additional columns to an existing select query, you can call AndSelect().
To select distinct rows, you can call Distinct(true).
app.DB().
Select("id", "avatar as image").
AndSelect("(firstName || ' ' || lastName) as fullName").
Distinct(true)
...  ##### From()
The From(...tables) method specifies which tables to select from (plain table names are automatically
quoted).
app.DB().
Select("table1.id", "table2.name").
From("table1", "table2")
...  ##### Join()
The Join(type, table, on) method specifies a JOIN clause. It takes 3 parameters:
-type - join type string like INNER JOIN, LEFT JOIN, etc.
-table - the name of the table to be joined
-on - optional dbx.Expression as an ON clause
For convenience, you can also use the shortcuts InnerJoin(table, on),
LeftJoin(table, on),
RightJoin(table, on) to specify INNER JOIN, LEFT JOIN and
RIGHT JOIN, respectively.
app.DB().
Select("users.*").
From("users").
InnerJoin("profiles", dbx.NewExp("profiles.user_id = users.id")).
Join("FULL OUTER JOIN", "department", dbx.NewExp("department.id = {:id}", dbx.Params{ "id": "someId" }))
...  ##### Where(), AndWhere(), OrWhere()
The Where(exp) method specifies the WHERE condition of the query.
You can also use AndWhere(exp) or OrWhere(exp) to append additional one or more
conditions to an existing WHERE clause.
Each where condition accepts a single dbx.Expression (see below for full list).
SELECT users.*
FROM users
WHERE id = "someId" AND
status = "public" AND
name like "%john%" OR
role = "manager" AND
fullTime IS TRUE AND
experience > 10
app.DB().
Select("users.*").
From("users").
Where(dbx.NewExp("id = {:id}", dbx.Params{ "id": "someId" })).
AndWhere(dbx.HashExp{"status": "public"}).
AndWhere(dbx.Like("name", "john")).
OrWhere(dbx.And(
dbx.HashExp{
"role":     "manager",
"fullTime": true,
dbx.NewExp("experience > {:exp}", dbx.Params{ "exp": 10 })
...  The following dbx.Expression methods are available:
dbx.Params to the expression.
dbx.NewExp("status = 'public'")
dbx.NewExp("total > {:min} AND total &lt; {:max}", dbx.Params{ "min": 10, "max": 30 })
-dbx.HashExp{k:v}
Generates a hash expression from a map whose keys are DB column names which need to be filtered according
to the corresponding values.
// slug = "example" AND active IS TRUE AND tags in ("tag1", "tag2", "tag3") AND parent IS NULL
dbx.HashExp{
"slug":   "example",
"active": true,
"tags":   []any{"tag1", "tag2", "tag3"},
"parent": nil,
-dbx.Not(exp)
Negates a single expression by wrapping it with NOT().
// NOT(status = 1)
dbx.Not(dbx.NewExp("status = 1"))
-dbx.And(...exps)
Creates a new expression by concatenating the specified ones with AND.
// (status = 1 AND username like "%john%")
dbx.And(
dbx.NewExp("status = 1"),
dbx.Like("username", "john"),
-dbx.Or(...exps)
Creates a new expression by concatenating the specified ones with OR.
// (status = 1 OR username like "%john%")
dbx.Or(
dbx.NewExp("status = 1"),
dbx.Like("username", "john")
-dbx.In(col, ...values)
Generates an IN expression for the specified column and the list of allowed values.
// status IN ("public", "reviewed")
dbx.In("status", "public", "reviewed")
-dbx.NotIn(col, ...values)
Generates an NOT IN expression for the specified column and the list of allowed values.
// status NOT IN ("public", "reviewed")
dbx.NotIn("status", "public", "reviewed")
-dbx.Like(col, ...values)
Generates a LIKE expression for the specified column and the possible strings that the
column should be like. If multiple values are present, the column should be like
all of them.
By default, each value will be surrounded by &quot;%&quot; to enable partial matching. Special
characters like &quot;%&quot;, &quot;\&quot;, &quot;_&quot; will also be properly escaped. You may call
Escape(...pairs) and/or Match(left, right) to change the default behavior.
// name LIKE "%test1%" AND name LIKE "%test2%"
dbx.Like("name", "test1", "test2")
// name LIKE "test1%"
dbx.Like("name", "test1").Match(false, true)
-dbx.NotLike(col, ...values)
Generates a NOT LIKE expression in similar manner as Like().
// name NOT LIKE "%test1%" AND name NOT LIKE "%test2%"
dbx.NotLike("name", "test1", "test2")
// name NOT LIKE "test1%"
dbx.NotLike("name", "test1").Match(false, true)
-dbx.OrLike(col, ...values)
This is similar to Like() except that the column must be one of the provided values, aka.
multiple values are concatenated with OR instead of AND.
// name LIKE "%test1%" OR name LIKE "%test2%"
dbx.OrLike("name", "test1", "test2")
// name LIKE "test1%" OR name LIKE "test2%"
dbx.OrLike("name", "test1", "test2").Match(false, true)
-dbx.OrNotLike(col, ...values)
This is similar to NotLike() except that the column must not be one of the provided
values, aka. multiple values are concatenated with OR instead of AND.
// name NOT LIKE "%test1%" OR name NOT LIKE "%test2%"
dbx.OrNotLike("name", "test1", "test2")
// name NOT LIKE "test1%" OR name NOT LIKE "test2%"
dbx.OrNotLike("name", "test1", "test2").Match(false, true)
-dbx.Exists(exp)
Prefix with EXISTS the specified expression (usually a subquery).
// EXISTS (SELECT 1 FROM users WHERE status = 'active')
dbx.Exists(dbx.NewExp("SELECT 1 FROM users WHERE status = 'active'"))
-dbx.NotExists(exp)
Prefix with NOT EXISTS the specified expression (usually a subquery).
// NOT EXISTS (SELECT 1 FROM users WHERE status = 'active')
dbx.NotExists(dbx.NewExp("SELECT 1 FROM users WHERE status = 'active'"))
-dbx.Between(col, from, to)
Generates a BETWEEN expression with the specified range.
// age BETWEEN 3 and 99
dbx.Between("age", 3, 99)
-dbx.NotBetween(col, from, to)
Generates a NOT BETWEEN expression with the specified range.
// age NOT BETWEEN 3 and 99
dbx.NotBetween("age", 3, 99)
##### OrderBy(), AndOrderBy()
The OrderBy(...cols) specifies the ORDER BY clause of the query.
A column name can contain &quot;ASC&quot; or &quot;DESC&quot; to indicate its ordering direction.
You can also use AndOrderBy(...cols) to append additional columns to an existing
ORDER BY clause.
app.DB().
Select("users.*").
From("users").
OrderBy("created ASC", "updated DESC").
AndOrderBy("title ASC")
...  ##### GroupBy(), AndGroupBy()
The GroupBy(...cols) specifies the GROUP BY clause of the query.
You can also use AndGroupBy(...cols) to append additional columns to an existing
GROUP BY clause.
app.DB().
Select("users.*").
From("users").
GroupBy("department", "level")
...  ##### Having(), AndHaving(), OrHaving()
The Having(exp) specifies the HAVING clause of the query.
Similarly to
Where(exp), it accept a single dbx.Expression (see all available expressions
listed above).
You can also use AndHaving(exp) or OrHaving(exp) to append additional one or
more conditions to an existing HAVING clause.
app.DB().
Select("users.*").
From("users").
GroupBy("department", "level").
Having(dbx.NewExp("sum(level) > {:sum}", dbx.Params{ sum: 10 }))
...  ##### Limit()
The Limit(number) method specifies the LIMIT clause of the query.
app.DB().
Select("users.*").
From("users").
Limit(30)
...  ##### Offset()
The Offset(number) method specifies the OFFSET clause of the query. Usually used
together with Limit(number).
app.DB().
Select("users.*").
From("users").
Offset(5).
Limit(30)
...  ### Transaction
To execute multiple queries in a transaction you can use
app.RunInTransaction(fn)
The DB operations are persisted only if the transaction returns nil.
It is safe to nest RunInTransaction calls as long as you use the callback&#39;s
txApp argument.
Inside the transaction function always use its txApp argument and not the original
app instance because we allow only a single writer/transaction at a time and it could
result in a deadlock.
To avoid performance issues, try to minimize slow/long running tasks such as sending emails,
connecting to external services, etc. as part of the transaction.
err := app.RunInTransaction(func(txApp core.App) error {
// update a record
record, err := txApp.FindRecordById("articles", "RECORD_ID")
if err != nil {
return err
record.Set("status", "active")
if err := txApp.Save(record); err != nil {
return err
return err
return nil

## 29.Extend with Go - Record operations
# Extend with Go - Record operations
Record operations The most common task when using PocketBase as framework probably would be querying and working with your
collection records.
You could find detailed documentation about all the supported Record model methods in
core.Record
but below are some examples with the most common ones.
### Set field value
// sets the value of a single record field
// (field type specific modifiers are also supported)
record.Set("title", "example")
record.Set("users+", "6jyr1y02438et52") // append to existing value
// populates a record from a data map
// (calls Set for each entry of the map)
record.Load(data)  ### Get field value
// retrieve a single record field value
// (field specific modifiers are also supported)
record.Get("someField")            // -> any (without cast)
record.GetBool("someField")        // -> cast to bool
record.GetString("someField")      // -> cast to string
record.GetInt("someField")         // -> cast to int
record.GetFloat("someField")       // -> cast to float64
record.GetDateTime("someField")    // -> cast to types.DateTime
record.GetStringSlice("someField") // -> cast to []string
// retrieve the new uploaded files
// (e.g. for inspecting and modifying the file(s) before save)
record.GetUnsavedFiles("someFileField")
// unmarshal a single "json" field value into the provided result
record.UnmarshalJSONField("someJSONField", &amp;result)
// retrieve a single or multiple expanded data
record.ExpandedOne("author")     // -> nil|*core.Record
record.ExpandedAll("categories") // -> []*core.Record
// export all the public safe record fields as map[string]any
record.PublicExport()  ### Auth accessors
record.IsSuperuser() // alias for record.Collection().Name == "_superusers"
record.Email()         // alias for record.Get("email")
record.SetEmail(email) // alias for record.Set("email", email)
record.Verified()         // alias for record.Get("verified")
record.SetVerified(false) // alias for record.Set("verified", false)
record.TokenKey()        // alias for record.Get("tokenKey")
record.SetTokenKey(key)  // alias for record.Set("tokenKey", key)
record.RefreshTokenKey() // alias for record.Set("tokenKey:autogenerate", "")
record.ValidatePassword(pass)
record.SetPassword(pass)   // alias for record.Set("password", pass)
record.SetRandomPassword() // sets cryptographically random 30 characters string as password  ### Copies
// returns a shallow copy of the current record model populated
// with its ORIGINAL db data state and everything else reset to the defaults
// (usually used for comparing old and new field values)
record.Original()
// returns a shallow copy of the current record model populated
// with its LATEST data state and everything else reset to the defaults
// (aka. no expand, no custom fields and with default visibility flags)
record.Fresh()
// returns a shallow copy of the current record model populated
// with its ALL collection and custom fields data, expand and visibility flags
record.Clone()  ### Hide/Unhide fields
Collection fields can be marked as &quot;Hidden&quot; from the Dashboard to prevent regular user access to the field
values.
Record models provide an option to further control the fields serialization visibility in addition to the
&quot;Hidden&quot; fields option using the
record.Hide(fieldNames...)
and
record.Unhide(fieldNames...)
methods.
Often the Hide/Unhide methods are used in combination with the OnRecordEnrich hook
invoked on every record enriching (list, view, create, update, realtime change, etc.). For example:
app.OnRecordEnrich("articles").BindFunc(func(e *core.RecordEnrichEvent) error {
// dynamically show/hide a record field depending on whether the current
// authenticated user has a certain "role" (or any other field constraint)
if e.RequestInfo.Auth == nil ||
(!e.RequestInfo.Auth.IsSuperuser() &amp;&amp; e.RequestInfo.Auth.GetString("role") != "staff") {
e.Record.Hide("someStaffOnlyField")
})   For custom fields, not part of the record collection schema, it is required to call explicitly
record.WithCustomData(true) to allow them in the public serialization.
### Fetch records
##### Fetch single record
All single record retrieval methods return nil and sql.ErrNoRows error if no record
is found.
// retrieve a single "articles" record by its id
record, err := app.FindRecordById("articles", "RECORD_ID")
// retrieve a single "articles" record by a single key-value pair
record, err := app.FindFirstRecordByData("articles", "slug", "test")
// retrieve a single "articles" record by a string filter expression
record, err := app.FindFirstRecordByFilter(
"articles",
"status = 'public' &amp;&amp; category = {:category}",
dbx.Params{ "category": "news" },
)  ##### Fetch multiple records
All multiple records retrieval methods return empty slice and nil error if no records are found.
// retrieve multiple "articles" records by their ids
records, err := app.FindRecordsByIds("articles", []string{"RECORD_ID1", "RECORD_ID2"})
// retrieve the total number of "articles" records in a collection with optional dbx expressions
totalPending, err := app.CountRecords("articles", dbx.HashExp{"status": "pending"})
// retrieve multiple "articles" records with optional dbx expressions
records, err := app.FindAllRecords("articles",
dbx.NewExp("LOWER(username) = {:username}", dbx.Params{"username": "John.Doe"}),
dbx.HashExp{"status": "pending"},
// retrieve multiple paginated "articles" records by a string filter expression
records, err := app.FindRecordsByFilter(
"articles",                                    // collection
"status = 'public' &amp;&amp; category = {:category}", // filter
"-published",                                   // sort
10,                                            // limit
0,                                             // offset
dbx.Params{ "category": "news" },              // optional filter params
)  ##### Fetch auth records
// retrieve a single auth record by its email
// retrieve a single auth record by JWT
// (you could also specify an optional list of accepted token types)
user, err := app.FindAuthRecordByToken("YOUR_TOKEN", core.TokenTypeAuth)  ##### Custom record query
In addition to the above query helpers, you can also create custom Record queries using
RecordQuery(collection)
method. It returns a SELECT DB builder that can be used with the same methods described in the
Database guide.
import (
"github.com/pocketbase/dbx"
"github.com/pocketbase/pocketbase/core"
func FindActiveArticles(app core.App) ([]*core.Record, error) {
records := []*core.Record{}
err := app.RecordQuery("articles").
AndWhere(dbx.HashExp{"status": "active"}).
OrderBy("published DESC").
Limit(10).
All(&amp;records)
if err != nil {
return nil, err
return records, nil
}  ### Create new record
##### Create new record programmatically
import (
"github.com/pocketbase/pocketbase/core"
"github.com/pocketbase/pocketbase/tools/filesystem"
collection, err := app.FindCollectionByNameOrId("articles")
if err != nil {
return err
record := core.NewRecord(collection)
record.Set("active", true)
// field type specific modifiers can also be used
record.Set("slug:autogenerate", "post-")
// new files must be one or a slice of *filesystem.File values
// note1: see all factories in https://pkg.go.dev/github.com/pocketbase/pocketbase/tools/filesystem#File
// note2: for reading files from a request event you can also use e.FindUploadedFiles("fileKey")
f1, _ := filesystem.NewFileFromPath("/local/path/to/file1.txt")
f2, _ := filesystem.NewFileFromBytes([]byte{"test content"}, "file2.txt")
record.Set("documents", []*filesystem.File{f1, f2, f3})
// validate and persist
// (use SaveNoValidate to skip fields validation)
err = app.Save(record);
if err != nil {
return err
}  ##### Intercept create request
import (
"github.com/pocketbase/pocketbase/core"
app.OnRecordCreateRequest("articles").BindFunc(func(e *core.RecordRequestEvent) error {
// ignore for superusers
if e.HasSuperuserAuth() {
// overwrite the submitted "status" field value
e.Record.Set("status", "pending")
// or you can also prevent the create event by returning an error
status := e.Record.GetString("status")
if (status != "pending" &amp;&amp;
// guest or not an editor
(e.Auth == nil || e.Auth.GetString("role") != "editor")) {
return e.BadRequestError("Only editors can set a status different from pending", nil)
})  ### Update existing record
##### Update existing record programmatically
record, err := app.FindRecordById("articles", "RECORD_ID")
if err != nil {
return err
// delete existing record files by specifying their file names
record.Set("documents-", []string{"file1_abc123.txt", "file3_abc123.txt"})
// append one or more new files to the already uploaded list
// note1: see all factories in https://pkg.go.dev/github.com/pocketbase/pocketbase/tools/filesystem#File
// note2: for reading files from a request event you can also use e.FindUploadedFiles("fileKey")
f1, _ := filesystem.NewFileFromPath("/local/path/to/file1.txt")
f2, _ := filesystem.NewFileFromBytes([]byte{"test content"}, "file2.txt")
record.Set("documents+", []*filesystem.File{f1, f2, f3})
// validate and persist
// (use SaveNoValidate to skip fields validation)
err = app.Save(record);
if err != nil {
return err
}  ##### Intercept update request
import (
"github.com/pocketbase/pocketbase/core"
app.OnRecordUpdateRequest("articles").Add(func(e *core.RecordRequestEvent) error {
// ignore for superusers
if e.HasSuperuserAuth() {
// overwrite the submitted "status" field value
e.Record.Set("status", "pending")
// or you can also prevent the update event by returning an error
status := e.Record.GetString("status")
if (status != "pending" &amp;&amp;
// guest or not an editor
(e.Auth == nil || e.Auth.GetString("role") != "editor")) {
return e.BadRequestError("Only editors can set a status different from pending", nil)
})  ### Delete record
record, err := app.FindRecordById("articles", "RECORD_ID")
if err != nil {
return err
err = app.Delete(record)
if err != nil {
return err
}  ### Transaction
To execute multiple queries in a transaction you can use
app.RunInTransaction(fn)
The DB operations are persisted only if the transaction returns nil.
It is safe to nest RunInTransaction calls as long as you use the callback&#39;s
txApp argument.
Inside the transaction function always use its txApp argument and not the original
app instance because we allow only a single writer/transaction at a time and it could
result in a deadlock.
To avoid performance issues, try to minimize slow/long running tasks such as sending emails,
connecting to external services, etc. as part of the transaction.
import (
"github.com/pocketbase/pocketbase/core"
titles := []string{"title1", "title2", "title3"}
collection, err := app.FindCollectionByNameOrId("articles")
if err != nil {
return err
// create new record for each title
app.RunInTransaction(func(txApp core.App) error {
for _, title := range titles {
record := core.NewRecord(collection)
record.Set("title", title)
if err := txApp.Save(record); err != nil {
return err
return nil
})  ### Programmatically expanding relations
To expand record relations programmatically you can use
app.ExpandRecord(record, expands, optFetchFunc)
for single or
app.ExpandRecords(records, expands, optFetchFunc)
for multiple records.
Once loaded, you can access the expanded relations via
record.ExpandedOne(relName)
record.ExpandedAll(relName) .
For example:
record, err := app.FindFirstRecordByData("articles", "slug", "lorem-ipsum")
if err != nil {
return err
// expand the "author" and "categories" relations
errs := app.ExpandRecord(record, []string{"author", "categories"}, nil)
if len(errs) > 0 {
return fmt.Errorf("failed to expand: %v", errs)
// print the expanded records
log.Println(record.ExpandedOne("author"))
log.Println(record.ExpandedAll("categories"))  ### Check if record can be accessed
To check whether a custom client request or user can access a single record, you can use the
app.CanAccessRecord(record, requestInfo, rule)
method.
Below is an example of creating a custom route to retrieve a single article and checking if the request
satisfy the View API rule of the record collection:
package main
import (
"log"
"net/http"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
app.OnServe().BindFunc(func(se *core.ServeEvent) error {
se.Router.GET("/articles/{slug}", func(e *core.RequestEvent) error {
slug := e.Request.PathValue("slug")
record, err := e.App.FindFirstRecordByData("articles", "slug", slug)
if err != nil {
return e.NotFoundError("Missing or invalid slug", err)
info, err := e.RequestInfo()
if err != nil {
return e.BadRequestError("Failed to retrieve request info", err)
canAccess, err := e.App.CanAccessRecord(record, info, record.Collection().ViewRule)
if !canAccess {
return e.ForbiddenError("", err)
return e.JSON(http.StatusOK, record)
if err := app.Start(); err != nil {
log.Fatal(err)
}  ### Generating and validating tokens
PocketBase Web APIs are fully stateless (aka. there are no sessions in the traditional sense) and an auth
record is considered authenticated if the submitted request contains a valid
Authorization: TOKEN
header
(see also Builtin auth middlewares and
Retrieving the current auth state from a route
If you want to issue and verify manually a record JWT (auth, verification, password reset, etc.), you
could do that using the record token type specific methods:
token, err := record.NewAuthToken()
token, err := record.NewVerificationToken()
token, err := record.NewPasswordResetToken()
token, err := record.NewEmailChangeToken(newEmail)
token, err := record.NewFileToken() // for protected files
token, err := record.NewStaticAuthToken(optCustomDuration) // nonrenewable auth token  Each token type has its own secret and the token duration is managed via its type related collection auth
option (the only exception is NewStaticAuthToken).
To validate a record token you can use the
app.FindAuthRecordByToken
method. The token related auth record is returned only if the token is not expired and its signature is valid.
Here is an example how to validate an auth token:
`record, err := app.FindAuthRecordByToken("YOUR_TOKEN", core.TokenTypeAuth)`

## 30.Extend with Go - Collection operations
# Extend with Go - Collection operations
Collection operations Collections are usually managed via the Dashboard interface, but there are some situations where you may
want to create or edit a collection programmatically (usually as part of a
DB migration). You can find all available Collection related operations
and methods in
core.App
and
core.Collection
, but below are listed some of the most common ones:
### Fetch collections
##### Fetch single collection
All single collection retrieval methods return nil and sql.ErrNoRows error if no
collection is found.
`collection, err := app.FindCollectionByNameOrId("example")`  ##### Fetch multiple collections
All multiple collections retrieval methods return empty slice and nil error if no collections
are found.
allCollections, err := app.FindAllCollections()
authAndViewCollections, err := app.FindAllCollections(core.CollectionTypeAuth, core.CollectionTypeView)  ##### Custom collection query
In addition to the above query helpers, you can also create custom Collection queries using
CollectionQuery()
method. It returns a SELECT DB builder that can be used with the same methods described in the
Database guide.
import (
"github.com/pocketbase/dbx"
"github.com/pocketbase/pocketbase/core"
func FindSystemCollections(app core.App) ([]*core.Collection, error) {
collections := []*core.Collection{}
err := app.CollectionQuery().
AndWhere(dbx.HashExp{"system": true}).
OrderBy("created DESC").
All(&amp;collections)
if err != nil {
return nil, err
return collections, nil
}  ### Collection properties
Id      string
Name    string
Type    string // "base", "view", "auth"
System  bool // !prevent collection rename, deletion and rules change of internal collections like _superusers
Fields  core.FieldsList
Indexes types.JSONArray[string]
Created types.DateTime
Updated types.DateTime
// CRUD rules
ListRule   *string
ViewRule   *string
CreateRule *string
UpdateRule *string
DeleteRule *string
// "view" type specific options
// (see https://github.com/pocketbase/pocketbase/blob/master/core/collection_model_view_options.go)
ViewQuery string
// "auth" type specific options
// (see https://github.com/pocketbase/pocketbase/blob/master/core/collection_model_auth_options.go)
AuthRule                   *string
ManageRule                 *string
AuthAlert                  core.AuthAlertConfig
OAuth2                     core.OAuth2Config
PasswordAuth               core.PasswordAuthConfig
MFA                        core.MFAConfig
OTP                        core.OTPConfig
AuthToken                  core.TokenConfig
PasswordResetToken         core.TokenConfig
EmailChangeToken           core.TokenConfig
VerificationToken          core.TokenConfig
FileToken                  core.TokenConfig
VerificationTemplate       core.EmailTemplate
ResetPasswordTemplate      core.EmailTemplate
ConfirmEmailChangeTemplate core.EmailTemplate  ### Field definitions
-core.BoolField
-core.NumberField
-core.TextField
-core.EmailField
-core.URLField
-core.EditorField
-core.DateField
-core.AutodateField
-core.SelectField
-core.FileField
-core.RelationField
-core.JSONField
-core.GeoPointField
### Create new collection
import (
"github.com/pocketbase/pocketbase/core"
"github.com/pocketbase/pocketbase/tools/types"
// core.NewAuthCollection("example")
// core.NewViewCollection("example")
collection := core.NewBaseCollection("example")
// set rules
collection.ViewRule = types.Pointer("@request.auth.id != ''")
collection.CreateRule = types.Pointer("@request.auth.id != '' &amp;&amp; @request.body.user = @request.auth.id")
collection.UpdateRule = types.Pointer(`
@request.auth.id != '' &amp;&amp;
user = @request.auth.id &amp;&amp;
(@request.body.user:isset = false || @request.body.user = @request.auth.id)
// add text field
collection.Fields.Add(&amp;core.TextField{
Name:     "title",
Required: true,
Max:      100,
// add relation field
usersCollection, err := app.FindCollectionByNameOrId("users")
if err != nil {
return err
collection.Fields.Add(&amp;core.RelationField{
Name:          "user",
Required:      true,
Max:           100,
CascadeDelete: true,
CollectionId:  usersCollection.Id,
// add autodate/timestamp fields (created/updated)
collection.Fields.Add(&amp;core.AutodateField{
Name:     "created",
OnCreate: true,
collection.Fields.Add(&amp;core.AutodateField{
Name:     "updated",
OnCreate: true,
OnUpdate: true,
// or: collection.Indexes = []string{"CREATE UNIQUE INDEX idx_example_user ON example (user)"}
collection.AddIndex("idx_example_user", true, "user", "")
// validate and persist
// (use SaveNoValidate to skip fields validation)
err = app.Save(collection)
if err != nil {
return err
}  ### Update existing collection
import (
"github.com/pocketbase/pocketbase/core"
"github.com/pocketbase/pocketbase/tools/types"
collection, err := app.FindCollectionByNameOrId("example")
if err != nil {
return err
// change rule
collection.DeleteRule = types.Pointer("@request.auth.id != ''")
// add new editor field
collection.Fields.Add(&amp;core.EditorField{
Name:     "description",
Required: true,
// change existing field
// (returns a pointer and direct modifications are allowed without the need of reinsert)
titleField := collection.Fields.GetByName("title").(*core.TextField)
titleField.Min = 10
// or: collection.Indexes = append(collection.Indexes, "CREATE INDEX idx_example_title ON example (title)")
collection.AddIndex("idx_example_title", false, "title", "")
// validate and persist
// (use SaveNoValidate to skip fields validation)
err = app.Save(collection)
if err != nil {
return err
}  ### Delete collection
collection, err := app.FindCollectionByNameOrId("example")
if err != nil {
return err
err = app.Delete(collection)
if err != nil {
return err

## 31.Extend with Go - Migrations
# Extend with Go - Migrations
Migrations PocketBase comes with a builtin DB and data migration utility, allowing you to version your DB structure,
create collections programmatically, initialize default settings, etc.
Because the migrations are regular Go functions, besides applying schema changes, they could be used also
to adjust existing data to fit the new schema or any other app specific logic that you want to run only
once.
And as a bonus, being .go files also ensure that the migrations will be embedded seamlessly in
your final executable.
### Quick setup
##### 0. Register the migrate command
You can find all available config options in the
migratecmd
subpackage.
// main.go
package main
import (
"log"
"strings"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/plugins/migratecmd"
// enable once you have at least one migration
// _ "yourpackage/migrations"
func main() {
app := pocketbase.New()
// loosely check if it was executed using "go run"
isGoRun := strings.HasPrefix(os.Args[0], os.TempDir())
migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
// enable auto creation of migration files when making collection changes in the Dashboard
// (the isGoRun check is to enable it only during development)
Automigrate: isGoRun,
if err := app.Start(); err != nil {
log.Fatal(err)
}  ##### 1. Create new migration
To create a new blank migration you can run migrate create.
// Since the "create" command makes sense only during development,
// it is expected the user to be in the app working directory
// and to be using "go run"
[root@dev app]$ go run . migrate create "your_new_migration"   // migrations/1655834400_your_new_migration.go
package migrations
import (
"github.com/pocketbase/pocketbase/core"
m "github.com/pocketbase/pocketbase/migrations"
func init() {
m.Register(func(app core.App) error {
// add up queries...
return nil
}, func(app core.App) error {
// add down queries...
return nil
}  The above will create a new blank migration file inside the default command migrations directory.
Each migration file should have a single m.Register(upFunc, downFunc) call.
In the migration file, you are expected to write your &quot;upgrade&quot; code in the upFunc callback.
The downFunc is optional and it should contain the &quot;downgrade&quot; operations to revert the
changes made by the upFunc.
Both callbacks accept a transactional core.App instance.
You can explore the
Database guide,
Collection operations and
Record operations
for more details how to interact with the database. You can also find
some examples further below in this guide.
##### 2. Load migrations
To make your application aware of the registered migrations, you have to import the above
migrations package in one of your main package files:
package main
import _ "yourpackage/migrations"
// ...  ##### 3. Run migrations
New unapplied migrations are automatically executed when the application server starts, aka. on
serve.
Alternatively, you can also apply new migrations manually by running migrate up.
To revert the last applied migration(s), you can run migrate down [number].
When manually applying or reverting migrations, the serve process needs to be restarted so
that it can refresh its cached collections state.
### Collections snapshot
The migrate collections command generates a full snapshot of your current collections
configuration without having to type it manually. Similar to the migrate create command, this
will generate a new migration file in the
migrations directory.
// Since the "collections" command makes sense only during development,
// it is expected the user to be in the app working directory
// and to be using "go run"
[root@dev app]$ go run . migrate collections  By default the collections snapshot is imported in extend mode, meaning that collections and
fields that don&#39;t exist in the snapshot are preserved. If you want the snapshot to delete
missing collections and fields, you can edit the generated file and change the last argument of
ImportCollectionsByMarshaledJSON method to true.
### Migrations history
All applied migration filenames are stored in the internal _migrations table.
During local development often you might end up making various collection changes to test different approaches.
When Automigrate is enabled this could lead in a migration history with unnecessary intermediate
steps that may not be wanted in the final migration history.
To avoid the clutter and to prevent applying the intermediate steps in production, you can remove (or
squash) the unnecessary migration files manually and then update the local migrations history by running:
`[root@dev app]$ go run . migrate history-sync`  The above command will remove any entry from the _migrations table that doesn&#39;t have a related
migration file associated with it.
// migrations/1687801090_set_pending_status.go
package migrations
import (
"github.com/pocketbase/pocketbase/core"
m "github.com/pocketbase/pocketbase/migrations"
// set a default "pending" status to all empty status articles
func init() {
m.Register(func(app core.App) error {
_, err := app.DB().NewQuery("UPDATE articles SET status = 'pending' WHERE status = ''").Execute()
return err
}, nil)
}  ##### Initialize default application settings
// migrations/1687801090_initial_settings.go
package migrations
import (
"github.com/pocketbase/pocketbase/core"
m "github.com/pocketbase/pocketbase/migrations"
func init() {
m.Register(func(app core.App) error {
settings := app.Settings()
// for all available settings fields you could check
// https://github.com/pocketbase/pocketbase/blob/develop/core/settings_model.go#L121-L130
settings.Meta.AppName = "test"
settings.Logs.MaxDays = 2
settings.Logs.LogAuthId = true
settings.Logs.LogIP = false
return app.Save(settings)
}, nil)
}  ##### Creating initial superuser
For all supported record methods, you can refer to
Record operations
You can also create the initial super user using the
./pocketbase superuser create EMAIL PASS
command.
// migrations/1687801090_initial_superuser.go
package migrations
import (
"github.com/pocketbase/pocketbase/core"
m "github.com/pocketbase/pocketbase/migrations"
func init() {
m.Register(func(app core.App) error {
superusers, err := app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
if err != nil {
return err
record := core.NewRecord(superusers)
// note: the values can be eventually loaded via os.Getenv(key)
// or from a special local config file
record.Set("password", "1234567890")
return app.Save(record)
}, func(app core.App) error { // optional revert operation
if record == nil {
return nil // probably already deleted
return app.Delete(record)
}  ##### Creating collection programmatically
For all supported collection methods, you can refer to
Collection operations
// migrations/1687801090_create_clients_collection.go
package migrations
import (
"github.com/pocketbase/pocketbase/core"
"github.com/pocketbase/pocketbase/tools/types"
m "github.com/pocketbase/pocketbase/migrations"
func init() {
m.Register(func(app core.App) error {
// init a new auth collection with the default system fields and auth options
collection := core.NewAuthCollection("clients")
// restrict the list and view rules for record owners
collection.ListRule = types.Pointer("id = @request.auth.id")
collection.ViewRule = types.Pointer("id = @request.auth.id")
// add extra fields in addition to the default ones
collection.Fields.Add(
&amp;core.TextField{
Name:     "company",
Required: true,
Max:      100,
&amp;core.URLField{
Name:        "website",
Presentable: true,
// disable password auth and enable OTP only
collection.PasswordAuth.Enabled = false
collection.OTP.Enabled = true
collection.AddIndex("idx_clients_company", false, "company", "")
return app.Save(collection)
}, func(app core.App) error { // optional revert operation
collection, err := app.FindCollectionByNameOrId("clients")
if err != nil {
return err
return app.Delete(collection)

## 32.Extend with Go - Jobs scheduling
# Extend with Go - Jobs scheduling
Jobs scheduling If you have tasks that need to be performed periodically, you could set up crontab-like jobs with the
builtin app.Cron() (it returns an app scoped
cron.Cron value)
The jobs scheduler is started automatically on app serve, so all you have to do is register a
handler with
app.Cron().Add(id, cronExpr, handler)
app.Cron().MustAdd(id, cronExpr, handler)
(the latter panic if the cron expression is not valid).
Each scheduled job runs in its own goroutine and must have:
-id - identifier for the scheduled job; could be used to replace or remove an existing
job
-cron expression - e.g. 0 0 * * * (
supports numeric list, steps, ranges or
macros )
-handler - the function that will be executed every time when the job runs
Here is one minimal example:
// main.go
package main
import (
"log"
"github.com/pocketbase/pocketbase"
func main() {
app := pocketbase.New()
// prints "Hello!" every 2 minutes
app.Cron().MustAdd("hello", "*/2 * * * *", func() {
log.Println("Hello!")
if err := app.Start(); err != nil {
log.Fatal(err)
}  To remove already registered cron job you can call
app.Cron().Remove(id)
All registered app level cron jobs can be also previewed and triggered from the
Dashboard > Settings > Crons section.
Keep in mind that the app.Cron() is also used for running the system scheduled jobs
like the logs cleanup or auto backups (the jobs id is in the format __pb*__) and
replacing these system jobs or calling RemoveAll()/Stop() could have unintended
side-effects.
If you want more advanced control you can initialize your own cron instance independent from the
application via cron.New().

## 33.Extend with Go - Sending emails
# Extend with Go - Sending emails
Sending emails PocketBase provides a simple abstraction for sending emails via the
app.NewMailClient() factory.
Depending on your configured mail settings (Dashboard &gt; Settings &gt; Mail settings) it will use the
sendmail command or a SMTP client.
### Send custom email
You can send your own custom email from anywhere within the app (hooks, middlewares, routes, etc.) by
using app.NewMailClient().Send(message). Here is an example of sending a custom email after
user registration:
// main.go
package main
import (
"log"
"net/mail"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
"github.com/pocketbase/pocketbase/tools/mailer"
func main() {
app := pocketbase.New()
app.OnRecordCreateRequest("users").BindFunc(func(e *core.RecordRequestEvent) error {
return err
message := &amp;mailer.Message{
From: mail.Address{
Address: e.App.Settings().Meta.SenderAddress,
Name:    e.App.Settings().Meta.SenderName,
To:      []mail.Address{{Address: e.Record.Email()}},
Subject: "YOUR_SUBJECT...",
HTML:    "YOUR_HTML_BODY...",
// bcc, cc, attachments and custom headers are also supported...
return e.App.NewMailClient().Send(message)
if err := app.Start(); err != nil {
log.Fatal(err)
}  ### Overwrite system emails
If you want to overwrite the default system emails for forgotten password, verification, etc., you can
adjust the default templates available from the
Dashboard &gt; Collections &gt; Edit collection &gt; Options
Alternatively, you can also apply individual changes by binding to one of the
mailer hooks. Here is an example of appending a Record
field value to the subject using the OnMailerRecordPasswordResetSend hook:
// main.go
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
func main() {
app := pocketbase.New()
app.OnMailerRecordPasswordResetSend("users").BindFunc(func(e *core.MailerRecordEvent) error {
// modify the subject
e.Message.Subject += (" " + e.Record.GetString("name"))
if err := app.Start(); err != nil {
log.Fatal(err)

## 34.Extend with Go - Rendering templates
Response 200:
{{template "placeholderName" .}}
Response 200:
{{block "placeholderName" .}}default...{{end}}
Response 200:
{{define "placeholderName"}}custom...{{end}}
# Extend with Go - Rendering templates
Rendering templates  ### Overview
A common task when creating custom routes or emails is the need of generating HTML output.
There are plenty of Go template-engines available that you can use for this, but often for simple cases
the Go standard library html/template package should work just fine.
To make it slightly easier to load template files concurrently and on the fly, PocketBase also provides a
thin wrapper around the standard library in the
github.com/pocketbase/pocketbase/tools/template
utility package.
import "github.com/pocketbase/pocketbase/tools/template"
data := map[string]any{"name": "John"}
html, err := template.NewRegistry().LoadFiles(
"views/base.html",
"views/partial1.html",
"views/partial2.html",
).Render(data)  The general flow when working with composed and nested templates is that you create &quot;base&quot; template(s)
The dot object (.) in the above represents the data passed to the templates
via the Render(data) method.
By default the templates apply contextual (HTML, JS, CSS, URI) auto escaping so the generated template
For more information about the template syntax please refer to the
html/template
and
text/template
package godocs.
Another great resource is also the Hashicorp&#39;s
Learn Go Template Syntax
tutorial.
### Example HTML page with layout
Consider the following app directory structure:
myapp/
views/
layout.html
hello.html
main.go  We define the content for layout.html as:
&lt;!DOCTYPE html>
&lt;html lang="en">
&lt;head>
&lt;title>{{block "title" .}}Default app title{{end}}&lt;/title>
&lt;/head>
&lt;body>
Header...
{{block "body" .}}
Default app body...
{{end}}
&lt;/body>
&lt;/html>  We define the content for hello.html as:
{{define "title"}}
Page 1
{{end}}
{{define "body"}}
&lt;p>Hello from {{.name}}&lt;/p>
{{end}}  Then to output the final page, we&#39;ll register a custom /hello/:name route:
// main.go
package main
import (
"log"
"net/http"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/core"
"github.com/pocketbase/pocketbase/tools/template"
func main() {
app := pocketbase.New()
app.OnServe().BindFunc(func(se *core.ServeEvent) error {
// this is safe to be used by multiple goroutines
// (it acts as store for the parsed templates)
registry := template.NewRegistry()
se.Router.GET("/hello/{name}", func(e *core.RequestEvent) error {
name := e.Request.PathValue("name")
html, err := registry.LoadFiles(
"views/layout.html",
"views/hello.html",
).Render(map[string]any{
"name": name,
if err != nil {
// or redirect to a dedicated 404 HTML page
return e.NotFoundError("", err)
return e.HTML(http.StatusOK, html)
if err := app.Start(); err != nil {
log.Fatal(err)

## 35.Extend with Go - Console commands
# Extend with Go - Console commands
Console commands You can register custom console commands using
app.RootCmd.AddCommand(cmd), where cmd is a
cobra command.
Here is an example:
package main
import (
"log"
"github.com/pocketbase/pocketbase"
"github.com/spf13/cobra"
func main() {
app := pocketbase.New()
app.RootCmd.AddCommand(&amp;cobra.Command{
Use: "hello",
Run: func(cmd *cobra.Command, args []string) {
log.Println("Hello world!")
if err := app.Start(); err != nil {
log.Fatal(err)
}  To run the command you can build your Go application and execute:
# or "go run main.go hello"
./myapp hello   Keep in mind that the console commands execute in their own separate app process and run
independently from the main serve command (aka. hook and realtime events between different
processes are not shared with one another).

## 36.Extend with Go - Realtime messaging
# Extend with Go - Realtime messaging
Realtime messaging By default PocketBase sends realtime events only for Record create/update/delete operations (and for the OAuth2 auth redirect), but you are free to send custom realtime messages to the connected clients via the
app.SubscriptionsBroker() instance.
app.SubscriptionsBroker().Clients()
returns all connected
subscriptions.Client
indexed by their unique connection id.
app.SubscriptionsBroker().ChunkedClients(size)
is similar but returns the result as a chunked slice allowing you to split the iteration across several goroutines
(usually combined with
errgroup
The current auth record associated with a client could be accessed through
client.Get(apis.RealtimeClientAuthKey)
Note that a single authenticated user could have more than one active realtime connection (aka.
multiple clients). This could happen for example when opening the same app in different tabs,
browsers, devices, etc.
Below you can find a minimal code sample that sends a JSON payload to all clients subscribed to the
&quot;example&quot; topic:
func notify(app core.App, subscription string, data any) error {
if err != nil {
return err
message := subscriptions.Message{
Name: subscription,
group := new(errgroup.Group)
chunks := app.SubscriptionsBroker().ChunkedClients(300)
for _, chunk := range chunks {
group.Go(func() error {
for _, client := range chunk {
if !client.HasSubscription(subscription) {
continue
client.Send(message)
return nil
return group.Wait()
err := notify(app, "example", map[string]any{"test": 123})
if err != nil {
return err
}  From the client-side, users can listen to the custom subscription topic by doing something like:
Dart  import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
await pb.realtime.subscribe('example', (e) => {
console.log(e)
})  import 'package:pocketbase/pocketbase.dart';
final pb = PocketBase('http://127.0.0.1:8090');
await pb.realtime.subscribe('example', (e) {
print(e)

## 37.Extend with Go - Filesystem
# Extend with Go - Filesystem
Filesystem PocketBase comes with a thin abstraction between the local filesystem and S3.
To configure which one will be used you can adjust the storage settings from
Dashboard &gt; Settings &gt; Files storage section.
The filesystem abstraction can be accessed programmatically via the
app.NewFilesystem()
method.
Below are listed some of the most common operations but you can find more details in the
filesystem
subpackage.
Always make sure to call Close() at the end for both the created filesystem instance and
the retrieved file readers to prevent leaking resources.
### Reading files
To retrieve the file content of a single stored file you can use
GetReader(key)
Note that file keys often contain a prefix (aka. the &quot;path&quot; to the file). For record
files the full key is
collectionId/recordId/filename.
To retrieve multiple files matching a specific prefix you can use
List(prefix)
The below code shows a minimal example how to retrieve a single record file and copy its content into a
bytes.Buffer.
if err != nil {
return err
// construct the full file key by concatenating the record storage path with the specific filename
avatarKey := record.BaseFilesPath() + "/" + record.GetString("avatar")
// initialize the filesystem
fsys, err := app.NewFilesystem()
if err != nil {
return err
defer fsys.Close()
// retrieve a file reader for the avatar key
r, err := fsys.GetReader(avatarKey)
if err != nil {
return err
defer r.Close()
// do something with the reader...
content := new(bytes.Buffer)
_, err = io.Copy(content, r)
if err != nil {
return err
}  ### Saving files
There are several methods to save (aka. write/upload) files depending on the available file content
source:
-Upload([]byte, key)
-UploadFile(*filesystem.File, key)
-UploadMultipart(*multipart.FileHeader, key)
Most users rarely will have to use the above methods directly because for collection records the file
persistence is handled transparently when saving the record model (it will also perform size and MIME type
validation based on the collection file field options). For example:
record, err := app.FindRecordById("articles", "RECORD_ID")
if err != nil {
return err
// Other available File factories
// - filesystem.NewFileFromBytes(data, name)
// - filesystem.NewFileFromURL(ctx, url)
// - filesystem.NewFileFromMultipart(mh)
f, err := filesystem.NewFileFromPath("/local/path/to/file")
// set new file (can be single *filesytem.File or multiple []*filesystem.File)
// (if the record has an old file it is automatically deleted on successful Save)
record.Set("yourFileField", f)
err = app.Save(record)
if err != nil {
return err
}  ### Deleting files
Files can be deleted from the storage filesystem using
Delete(key)
because for collection records the file deletion is handled transparently when removing the existing filename
from the record model (this also ensures that the db entry referencing the file is also removed). For example:
record, err := app.FindRecordById("articles", "RECORD_ID")
if err != nil {
return err
// if you want to "reset" a file field (aka. deleting the associated single or multiple files)
// you can set it to nil
record.Set("yourFileField", nil)
// OR if you just want to remove individual file(s) from a multiple file field you can use the "-" modifier
// (the value could be a single filename string or slice of filename strings)
record.Set("yourFileField-", "example_52iWbGinWd.txt")
err = app.Save(record)
if err != nil {
return err

## 38.Extend with Go - Logging
# Extend with Go - Logging
Logging app.Logger() provides access to a standard slog.Logger implementation that
writes any logs into the database so that they can be later explored from the PocketBase
Dashboard &gt; Logs section.
For better performance and to minimize blocking on hot paths, logs are written with debounce and
on batches:
-3 seconds after the last debounced log write
-when the batch threshold is reached (currently 200)
-right before app termination to attempt saving everything from the existing logs queue
### Log methods
All standard
slog.Logger
methods are available but below is a list with some of the most notable ones.
##### Debug(message, attrs...)
app.Logger().Debug("Debug message!")
app.Logger().Debug(
"Debug message with attributes!",
"name", "John Doe",
"id", 123,
)  ##### Info(message, attrs...)
app.Logger().Info("Info message!")
app.Logger().Info(
"Info message with attributes!",
"name", "John Doe",
"id", 123,
)  ##### Warn(message, attrs...)
app.Logger().Warn("Warning message!")
app.Logger().Warn(
"Warning message with attributes!",
"name", "John Doe",
"id", 123,
)  ##### Error(message, attrs...)
app.Logger().Error("Error message!")
app.Logger().Error(
"Error message with attributes!",
"id", 123,
"error", err,
)  ##### With(attrs...)
With(atrs...) creates a new local logger that will &quot;inject&quot; the specified attributes with each
following log.
l := app.Logger().With("total", 123)
// results in log with data {"total": 123}
l.Info("message A")
// results in log with data {"total": 123, "name": "john"}
l.Info("message B", "name", "john")  ##### WithGroup(name)
WithGroup(name) creates a new local logger that wraps all logs attributes under the specified
group name.
l := app.Logger().WithGroup("sub")
// results in log with data {"sub": { "total": 123 }}
l.Info("message A", "total", 123)  ### Logs settings
You can control various log settings like logs retention period, minimal log level, request IP logging,
etc. from the logs settings panel:
### Custom log queries
The logs are usually meant to be filtered from the UI but if you want to programmatically retrieve and
filter the stored logs you can make use of the
app.LogQuery() query builder method. For example:
logs := []*core.Log{}
// see https://pocketbase.io/docs/go-database/#query-builder
err := app.LogQuery().
// target only debug and info logs
AndWhere(dbx.In("level", -4, 0).
// the data column is serialized json object and could be anything
AndWhere(dbx.NewExp("json_extract(data, '$.type') = 'request'")).
OrderBy("created DESC").
Limit(100).
All(&amp;logs)  ### Intercepting logs write
If you want to modify the log data before persisting in the database or to forward it to an external
system, then you can listen for changes of the _logs table by attaching to the
base model hooks. For example:
app.OnModelCreate(core.LogsTableName).BindFunc(func(e *core.ModelEvent) error {
l := e.Model.(*core.Log)
fmt.Println(l.Id)
fmt.Println(l.Created)
fmt.Println(l.Level)
fmt.Println(l.Message)
fmt.Println(l.Data)

## 39.Extend with Go - Testing
GET /my/hello
# Extend with Go - Testing
`GET /my/hello`
Testing PocketBase exposes several test mocks and stubs (eg. tests.TestApp,
tests.ApiScenario, tests.MockMultipartData, etc.) to help you write unit and
integration tests for your app.
You could find more information in the
github.com/pocketbase/pocketbase/tests
sub package, but here is a simple example.
### 1. Setup
Let&#39;s say that we have a custom API route GET /my/hello that requires superuser authentication:
// main.go
package main
import (
"log"
"net/http"
"github.com/pocketbase/pocketbase"
"github.com/pocketbase/pocketbase/apis"
"github.com/pocketbase/pocketbase/core"
func bindAppHooks(app core.App) {
app.OnServe().BindFunc(func(se *core.ServeEvent) error {
se.Router.Get("/my/hello", func(e *core.RequestEvent) error {
return e.JSON(http.StatusOK, "Hello world!")
}).Bind(apis.RequireSuperuserAuth())
func main() {
app := pocketbase.New()
bindAppHooks(app)
if err := app.Start(); err != nil {
log.Fatal(err)
}  ### 2. Prepare test data
Now we have to prepare our test/mock data. There are several ways you can approach this, but the easiest
one would be to start your application with a custom test_pb_data directory, e.g.:
`./pocketbase serve --dir="./test_pb_data" --automigrate=0`  Go to your browser and create the test data via the Dashboard (both collections and records). Once
completed you can stop the server (you could also commit test_pb_data to your repo).
### 3. Integration test
To test the example endpoint, we want to:
-ensure it handles only GET requests
-ensure that it can be accessed only by superusers
-check if the response body is properly set
Below is a simple integration test for the above test cases. We&#39;ll also use the test data created in the
// main_test.go
package main
import (
"net/http"
"testing"
"github.com/pocketbase/pocketbase/core"
"github.com/pocketbase/pocketbase/tests"
const testDataDir = "./test_pb_data"
func generateToken(collectionNameOrId string, email string) (string, error) {
app, err := tests.NewTestApp(testDataDir)
if err != nil {
return "", err
defer app.Cleanup()
record, err := app.FindAuthRecordByEmail(collectionNameOrId, email)
if err != nil {
return "", err
return record.NewAuthToken()
func TestHelloEndpoint(t *testing.T) {
if err != nil {
t.Fatal(err)
if err != nil {
t.Fatal(err)
// set up the test ApiScenario app instance
setupTestApp := func(t testing.TB) *tests.TestApp {
testApp, err := tests.NewTestApp(testDataDir)
if err != nil {
t.Fatal(err)
// no need to cleanup since scenario.Test() will do that for us
// defer testApp.Cleanup()
bindAppHooks(testApp)
return testApp
scenarios := []tests.ApiScenario{
Name:            "try with different http method, e.g. POST",
Method:          http.MethodPost,
URL:             "/my/hello",
ExpectedStatus:  405,
ExpectedContent: []string{"\"data\":{}"},
TestAppFactory:  setupTestApp,
Name:            "try as guest (aka. no Authorization header)",
Method:          http.MethodGet,
URL:             "/my/hello",
ExpectedStatus:  401,
ExpectedContent: []string{"\"data\":{}"},
TestAppFactory:  setupTestApp,
Name:   "try as authenticated app user",
Method: http.MethodGet,
URL:    "/my/hello",
Headers: map[string]string{
"Authorization": recordToken,
ExpectedStatus:  401,
ExpectedContent: []string{"\"data\":{}"},
TestAppFactory:  setupTestApp,
Name:   "try as authenticated superuser",
Method: http.MethodGet,
URL:    "/my/hello",
Headers: map[string]string{
"Authorization": superuserToken,
ExpectedStatus:  200,
ExpectedContent: []string{"Hello world!"},
TestAppFactory:  setupTestApp,
for _, scenario := range scenarios {
scenario.Test(t)

## 40.Extend with Go - Miscellaneous
# Extend with Go - Miscellaneous
Miscellaneous  ### app.Store()
app.Store()
returns a concurrent-safe application memory store that you can use to store anything for the duration of the
application process (e.g. cache, config flags, etc.).
You can find more details about the available store methods in the
store.Store
documentation but the most commonly used ones are Get(key), Set(key, value) and
GetOrSet(key, setFunc).
app.Store().Set("example", 123)
v1 := app.Store().Get("example").(int) // 123
v2 := app.Store().GetOrSet("example2", func() any {
// this setter is invoked only once unless "example2" is removed
// (e.g. suitable for instantiating singletons)
return 456
}).(int) // 456   Keep in mind that the application store is also used internally usually with pb*
prefixed keys (e.g. the collections cache is stored under the pbAppCachedCollections
key) and changing these system keys or calling RemoveAll()/Reset() could
have unintended side-effects.
If you want more advanced control you can initialize your own store independent from the
application instance via
store.New[K, T](nil).
### Security helpers
Below are listed some of the most commonly used security helpers but you can find detailed
documentation for all available methods in the
security
subpackage.
##### Generating random strings
secret := security.RandomString(10) // e.g. a35Vdb10Z4
secret := security.RandomStringWithAlphabet(5, "1234567890") // e.g. 33215  ##### Compare strings with constant time
`isEqual := security.Equal(hash1, hash2)`  ##### AES Encrypt/Decrypt
// must be random 32 characters string
const key = "sqAEm1v5NI3XyMKRBJdHCDANfwj3hqZ9"
encrypted, err := security.Encrypt([]byte("test"), key)
if err != nil {
return err
decrypted := security.Decrypt(encrypted, key) // []byte("test")

## 41.Extend with Go - Record proxy
# Extend with Go - Record proxy
Record proxy The available core.Record and its helpers
are usually the recommended way to interact with your data, but in case you want a typed access to your record
fields you can create a helper struct that embeds
core.BaseRecordProxy (which implements the core.RecordProxy interface) and define your collection fields as
getters and setters.
By implementing the core.RecordProxy interface you can use your custom struct as part of a
RecordQuery result like a regular record model. In addition, every DB change through the proxy
struct will trigger the corresponding record validations and hooks. This ensures that other parts of your app,
including 3rd party plugins, that don&#39;t know or use your custom struct will still work as expected.
Below is a sample Article record proxy implementation:
// article.go
package main
import (
"github.com/pocketbase/pocketbase/core"
"github.com/pocketbase/pocketbase/tools/types"
// ensures that the Article struct satisfy the core.RecordProxy interface
var _ core.RecordProxy = (*Article)(nil)
type Article struct {
core.BaseRecordProxy
func (a *Article) Title() string {
return a.GetString("title")
func (a *Article) SetTitle(title string) {
a.Set("title", title)
func (a *Article) Slug() string {
return a.GetString("slug")
func (a *Article) SetSlug(slug string) {
a.Set("slug", slug)
func (a *Article) Created() types.DateTime {
return a.GetDateTime("created")
func (a *Article) Updated() types.DateTime {
return a.GetDateTime("updated")
}  Accessing and modifying the proxy records is the same as for the regular records. Continuing with the
above Article example:
func FindArticleBySlug(app core.App, slug string) (*Article, error) {
article := &amp;Article{}
err := app.RecordQuery("articles").
AndWhere(dbx.NewExp("LOWER(slug)={:slug}", dbx.Params{
"slug": strings.ToLower(slug), // case insensitive match
Limit(1).
One(article)
if err != nil {
return nil, err
return article, nil
article, err := FindArticleBySlug(app, "example")
if err != nil {
return err
// change the title
// persist the change while also triggering the original record validations and hooks
err = app.Save(article)
if err != nil {
return err
}  If you have an existing *core.Record value you can also load it into your proxy using the
SetProxyRecord method:
// fetch regular record
record, err := app.FindRecordById("articles", "RECORD_ID")
if err != nil {
return err
// load into proxy
article := &amp;Article{}
article.SetProxyRecord(record)
