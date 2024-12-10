package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

func main() {
	// Parse command line arguments
	account := flag.String("account", "", "Facebook account")
	password := flag.String("password", "", "Facebook password")
	groupID := flag.String("group", "817620721658179", "Facebook group ID")
	postLimit := flag.Int("limit", 10, "Number of posts to scan")
	flag.Parse()

	if *account == "" || *password == "" {
		log.Fatal("Account and password are required. Use -account and -password flags")
	}

	// Configure Chrome
	opts := []selenium.ServiceOption{}
	caps := selenium.Capabilities{"browserName": "chrome"}
	chromeCaps := chrome.Capabilities{
		Args: []string{
			"--no-sandbox",
			"--disable-dev-shm-usage",
			"--disable-notifications", // Block notifications
			"--start-maximized",
			// "--headless",
		},
	}
	caps.AddChrome(chromeCaps)

	// Start Chrome
	driver, err := initializeDriver(opts, caps)
	if err != nil {
		log.Fatal("Failed to initialize driver:", err)
	}
	defer driver.Quit()

	// Login to Facebook
	if err := loginToFacebook(driver, *account, *password); err != nil {
		log.Fatal("Login failed:", err)
	}

	// Keywords to search for
	keywords := []string{
		"冰箱", "手錶", "手套", "套房", "票", 
		"GeForce", "BTS", "Airpod",
		// "二手", "販售", "售", "賣", // Added more relevant keywords
	}

	// Navigate to group and scan posts
	groupURL := fmt.Sprintf("https://www.facebook.com/groups/%s", *groupID)
	if err := scanGroupPosts(driver, groupURL, keywords, *postLimit); err != nil {
		log.Fatal("Failed to scan group:", err)
	}
}

func initializeDriver(opts []selenium.ServiceOption, caps selenium.Capabilities) (selenium.WebDriver, error) {
	service, err := selenium.NewChromeDriverService("chromedriver", 9515, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to start ChromeDriver: %v", err)
	}

	driver, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", 9515))
	if err != nil {
		service.Stop()
		return nil, fmt.Errorf("failed to create driver: %v", err)
	}

	return driver, nil
}

func loginToFacebook(driver selenium.WebDriver, account, password string) error {
	if err := driver.Get("https://www.facebook.com"); err != nil {
		return err
	}

	time.Sleep(2 * time.Second) // Wait for page load

	// Login
	if email, err := driver.FindElement(selenium.ByCSSSelector, "input[name='email']"); err == nil {
		email.SendKeys(account)
	} else {
		return fmt.Errorf("couldn't find email input: %v", err)
	}

	if pass, err := driver.FindElement(selenium.ByCSSSelector, "input[name='pass']"); err == nil {
		pass.SendKeys(password)
	} else {
		return fmt.Errorf("couldn't find password input: %v", err)
	}

	if loginBtn, err := driver.FindElement(selenium.ByCSSSelector, "button[name='login']"); err == nil {
		loginBtn.Click()
	} else {
		return fmt.Errorf("couldn't find login button: %v", err)
	}

	time.Sleep(30 * time.Second) // Wait for login
	return nil
}

func scanGroupPosts(driver selenium.WebDriver, groupURL string, keywords []string, postLimit int) error {
    if err := driver.Get(groupURL); err != nil {
		return err
	}
	time.Sleep(5 * time.Second)
	clickNewPost(driver)
	postsFound := 0
	lastPostCount := 0
	attempts := 0
	maxAttempts := 5

    for postsFound < postLimit && attempts < maxAttempts {
        driver.ExecuteScript("window.scrollTo(0, document.body.scrollHeight)", nil)
        expandPosts(driver)
		
        time.Sleep(4 * time.Second)
		posts, _ := driver.FindElements(selenium.ByCSSSelector, "div.x1yztbdb:not([aria-hidden='true'])")

        // posts, err := driver.FindElements(selenium.ByCSSSelector, "div[role='article']")
        // if err != nil || len(posts) == 0 {
        //     posts, _ = driver.FindElements(selenium.ByCSSSelector, "div.x1yztbdb:not([aria-hidden='true'])")
        // }

        if len(posts) == lastPostCount {
            attempts++
        } else {
            lastPostCount = len(posts)
            attempts = 0
        }

        for _, post := range posts {
            text, err := post.Text()
            if err != nil {
                continue
            }

            // Convert both text and keyword to []rune to handle Chinese characters properly
            postText := []rune(text)
            if len(postText) < 5 {
                continue
            }

            // Check for keywords with proper Chinese character handling
            for _, keyword := range keywords {
                keywordRunes := []rune(keyword)
                if strings.Contains(string(postText), string(keywordRunes)) {
                    fmt.Printf("\n=== 找到包含 '%s' 的貼文 ===\n", keyword)
                    fmt.Printf("貼文內容:\n%s\n", text)
                    fmt.Println("=====================================")
                    
                    // Try to get post timestamp and URL
                    if links, err := post.FindElements(selenium.ByCSSSelector, "a[href*='/groups/']"); err == nil && len(links) > 0 {
                        if href, err := links[0].GetAttribute("href"); err == nil {
                            fmt.Printf("貼文連結: %s\n", href)
                        }
                    }
                    
                    postsFound++
                    break
                }
            }

            if postsFound >= postLimit {
                break
            }
        }
    }

    fmt.Printf("\n掃描完成 共找到 %d 則符合關鍵字的貼文。\n", postsFound)
    return nil
}

func expandPosts(driver selenium.WebDriver) {
	// Click "See More" buttons
	if seeMoreBtns, err := driver.FindElements(selenium.ByXPATH, "//div[contains(text(),'查看更多')]"); err == nil {
		for _, btn := range seeMoreBtns {
			if err := btn.Click(); err == nil {
				time.Sleep(10000 * time.Millisecond)
			}
		}
	}
	time.Sleep(3 * time.Second)
}

func clickNewPost(driver selenium.WebDriver) {
	// if expandbutton,err := driver.FindElement(selenium.ByCSSSelector,"//*[@id='mount_0_0_PM']/div/div[1]/div/div[3]/div/div/div[1]/div[1]/div[4]/div/div[2]/div/div/div/div[2]/div[2]/div[1]/div/div/div/div/div/div/div/span/div/div[2]/div/div[2]/div/svg"); err == nil {
	// 	expandbutton.Click()
	// }
	if expandbutton, err := driver.FindElement(selenium.ByXPATH, "//span[contains(text(),'最相關')]"); err == nil {
		fmt.Println("找到最相關")
		expandbutton.Click()
	}
	if newPostBtn, err := driver.FindElement(selenium.ByXPATH, "//span[contains(text(),'新貼文')]"); err == nil {
		fmt.Println("找到新貼文")
		newPostBtn.Click()
	}
	time.Sleep(3 * time.Second)
}