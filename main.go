package main

import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "log"
    "net/http"
    "os"
    "time"
)

var db *sql.DB

var (
    usersLoggedInRole = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Namespace: "sapawarga",
            Name:      "users_loggedin_role",
            Help:      "Logged in users by roles",
        },[]string{
            "role",
        })

    usersLoggedInArea = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Namespace: "sapawarga",
            Name:      "users_loggedin_area",
            Help:      "Logged in users by area",
        },[]string{
            "kabkota",
        })

    usersRecentActiveArea = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Namespace: "sapawarga",
            Name:      "users_recent_active",
            Help:      "Recent active users by area",
        },[]string{
            "kabkota",
        })

    usersDailyActiveArea = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Namespace: "sapawarga",
            Name:      "users_active_daily",
            Help:      "Daily active users by area",
        },[]string{
            "kabkota",
        })

    usersWeeklyActiveArea = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Namespace: "sapawarga",
            Name:      "users_active_weekly",
            Help:      "Weekly active users by area",
        },[]string{
            "kabkota",
        })

    usersMonthlyActiveArea = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Namespace: "sapawarga",
            Name:      "users_active_monthly",
            Help:      "Monthly active users by area",
        },[]string{
            "kabkota",
        })

    usersPostArea = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Namespace: "sapawarga",
            Name:      "users_posts",
            Help:      "Recent user posts by area",
        },[]string{
            "kabkota",
        })
)

func main() {
    db = connect()

    http.Handle("/metrics", promhttp.Handler())

    prometheus.MustRegister(usersLoggedInRole)
    prometheus.MustRegister(usersLoggedInArea)
    prometheus.MustRegister(usersRecentActiveArea)
    prometheus.MustRegister(usersDailyActiveArea)
    prometheus.MustRegister(usersWeeklyActiveArea)
    prometheus.MustRegister(usersMonthlyActiveArea)
    prometheus.MustRegister(usersPostArea)

    go func() {
        for {
            watchRecentActiveUsers()

            time.Sleep(time.Minute * 1)
        }
    }()

    go func() {
        for {
            watchLoggedInUsersRoles()
            watchLoggedInUsersArea()
            watchDailyActiveUsers()
            watchWeeklyActiveUsers()
            watchMonthlyActiveUsers()
            watchUsersPostArea()

            time.Sleep(time.Minute * 5)
        }
    }()

    log.Fatal(http.ListenAndServe(":8000", nil))
}

func watchLoggedInUsersRoles() {
    var countStaffProv int
    var countStaffKabkota int
    var countStaffKecamatan int
    var countStaffKelurahan int
    var countRW int

    usersLoggedInRole.Reset()

    // --- Staff Prov Users
    err := db.QueryRow("select count(*) from user where role = 90 and status = 10 and last_login_at is not null;").Scan(&countStaffProv)

    if err != nil {
        log.Fatal(err)
    }

    // --- Staff Kabkota Users
    err = db.QueryRow("select count(*) from user where role = 80 and status = 10 and last_login_at is not null;").Scan(&countStaffKabkota)

    if err != nil {
        log.Fatal(err)
    }

    // --- Staff Kabkota Users
    err = db.QueryRow("select count(*) from user where role = 70 and status = 10 and last_login_at is not null;").Scan(&countStaffKecamatan)

    if err != nil {
        log.Fatal(err)
    }

    // --- Staff Kabkota Users
    err = db.QueryRow("select count(*) from user where role = 60 and status = 10 and last_login_at is not null;").Scan(&countStaffKelurahan)

    if err != nil {
        log.Fatal(err)
    }

    // --- RW Users
    err = db.QueryRow("select count(*) from user where role = 50 and status = 10 and last_login_at is not null;").Scan(&countRW)

    if err != nil {
        log.Fatal(err)
    }

    // log.Printf("users_logged_in_all %d", count)

    usersLoggedInRole.WithLabelValues("staffprov").Set(float64(countStaffProv))
    usersLoggedInRole.WithLabelValues("staffkabkota").Set(float64(countStaffKabkota))
    usersLoggedInRole.WithLabelValues("staffkecamatan").Set(float64(countStaffKecamatan))
    usersLoggedInRole.WithLabelValues("staffkelurahan").Set(float64(countStaffKelurahan))
    usersLoggedInRole.WithLabelValues("rw").Set(float64(countRW))
}

func watchLoggedInUsersArea() {
    var kabkota string
    var count int

    usersLoggedInArea.Reset()

    rows, _ := db.Query(`SELECT b.name, count(*) FROM user a join areas b on a.kabkota_id = b.id WHERE role = 50 and a.status = 10 and last_login_at is not null GROUP BY a.kabkota_id`)

    defer rows.Close()

    for rows.Next() {
        err := rows.Scan(&kabkota, &count)

        if err != nil {
            log.Fatal(err)
        }

        usersLoggedInArea.WithLabelValues(kabkota).Set(float64(count))
    }
}

func watchRecentActiveUsers() {
    var kabkota string
    var count int

    usersRecentActiveArea.Reset()

    rows, _ := db.Query(`SELECT b.name, count(*) FROM user a join areas b on a.kabkota_id = b.id WHERE a.role = 50 && a.last_access_at >= DATE_SUB(NOW(), INTERVAL 5 MINUTE) GROUP BY a.kabkota_id`)

    defer rows.Close()

    for rows.Next() {
        err := rows.Scan(&kabkota, &count)

        if err != nil {
            log.Fatal(err)
        }

        usersRecentActiveArea.WithLabelValues(kabkota).Set(float64(count))
    }
}

func watchDailyActiveUsers() {
    var kabkota string
    var count int

    usersDailyActiveArea.Reset()

    rows, _ := db.Query(`SELECT b.name, count(*) FROM user a join areas b on a.kabkota_id = b.id WHERE a.role = 50 && DATE(a.last_access_at) = DATE(NOW()) GROUP BY a.kabkota_id`)

    defer rows.Close()

    for rows.Next() {
        err := rows.Scan(&kabkota, &count)

        if err != nil {
            log.Fatal(err)
        }

        usersDailyActiveArea.WithLabelValues(kabkota).Set(float64(count))
    }
}

func watchWeeklyActiveUsers() {
    var kabkota string
    var count int

    usersWeeklyActiveArea.Reset()

    rows, _ := db.Query(`SELECT b.name, count(*) FROM user a JOIN areas b ON a.kabkota_id = b.id WHERE a.role = 50 && YEARWEEK(last_access_at, 1) = YEARWEEK(CURDATE(), 1) GROUP BY a.kabkota_id`)

    defer rows.Close()

    for rows.Next() {
        err := rows.Scan(&kabkota, &count)

        if err != nil {
            log.Fatal(err)
        }

        usersWeeklyActiveArea.WithLabelValues(kabkota).Set(float64(count))
    }
}

func watchMonthlyActiveUsers() {
    var kabkota string
    var count int

    usersMonthlyActiveArea.Reset()

    rows, _ := db.Query(`SELECT b.name, count(*) FROM user a JOIN areas b ON a.kabkota_id = b.id WHERE a.role = 50 && last_access_at >= DATE_FORMAT(NOW() ,'%Y-%m-01') GROUP BY a.kabkota_id`)

    defer rows.Close()

    for rows.Next() {
        err := rows.Scan(&kabkota, &count)

        if err != nil {
            log.Fatal(err)
        }

        usersMonthlyActiveArea.WithLabelValues(kabkota).Set(float64(count))
    }
}

func watchUsersPostArea() {
    var kabkota string
    var count int

    usersPostArea.Reset()

    rows, _ := db.Query(`SELECT c.name, count(*) FROM user_posts a JOIN user b ON a.created_by = b.id JOIN areas c ON b.kabkota_id = c.id GROUP BY b.kabkota_id;`)

    defer rows.Close()

    for rows.Next() {
        err := rows.Scan(&kabkota, &count)

        if err != nil {
            log.Fatal(err)
        }

        usersPostArea.WithLabelValues(kabkota).Set(float64(count))
    }
}

func connect() *sql.DB {
    dbHost := os.Getenv("DB_HOST")
    dbPort := os.Getenv("DB_PORT")
    dbName := os.Getenv("DB_NAME")
    dbUser := os.Getenv("DB_USER")

    db, err := sql.Open("mysql", dbUser+"@tcp("+dbHost+":"+dbPort+")/"+dbName)

    if err != nil {
        log.Fatal(err)
    }

    return db
}
