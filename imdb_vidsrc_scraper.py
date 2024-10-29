import requests
from bs4 import BeautifulSoup
import webbrowser

def search_imdb(query):
    url = f"https://www.imdb.com/find?q={query}&s=tt&ttype=tv,movie"
    response = requests.get(url)
    soup = BeautifulSoup(response.text, 'html.parser')
    
    results = []
    for item in soup.select('.findResult'):
        title = item.select_one('.result_text').text.strip()
        link = item.select_one('a')['href']
        imdb_id = link.split('/')[2]
        results.append((title, imdb_id))
    
    return results

def main():
    query = input("Enter the name of the show or movie you want to search for: ")
    results = search_imdb(query)
    
    if not results:
        print("No results found.")
        return
    
    print("\nSearch Results:")
    for i, (title, imdb_id) in enumerate(results, 1):
        print(f"{i}. {title} (IMDb ID: {imdb_id})")
    
    choice = int(input("\nEnter the number of the show/movie you want to watch: ")) - 1
    selected_title, selected_id = results[choice]
    
    print(f"\nYou selected: {selected_title}")
    print(f"IMDb ID: {selected_id}")
    
    vidsrc_url = f"https://vidsrc.cc/v2/embed/movie/{selected_id}"
    print(f"\nVidSrc URL: {vidsrc_url}")
    
    open_browser = input("Do you want to open this URL in your browser? (y/n): ").lower()
    if open_browser == 'y':
        webbrowser.open(vidsrc_url)

if __name__ == "__main__":
    main()
