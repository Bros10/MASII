from flask import Flask, render_template, request, redirect
from flask_sqlalchemy import SQLAlchemy

app = Flask(__name__)
app.config['SQLALCHEMY_DATABASE_URI'] = 'sqlite:///users.db'
db = SQLAlchemy(app)

class User(db.Model):
    id = db.Column(db.Integer, primary_key=True)
    username = db.Column(db.String(80), unique=True, nullable=False)

    def __repr__(self):
        return '<User %r>' % self.username

db.create_all()

@app.route('/')
def index():
    return render_template('index.html')

@app.route('/email')
def email():
    return render_template('email.html')

@app.route('/secret')
def secret():
    return "This is a secret page."

@app.route('/add_user/<username>')
def add_user(username):
    new_user = User(username=username)
    db.session.add(new_user)
    db.session.commit()
    return f"User '{username}' added successfully."

@app.route('/about')
def about():
    return render_template('about.html')

@app.route('/staff')
def staff():
    return render_template('staff.html')

@app.route('/gallery')
def gallery():
    return render_template('gallery.html')

@app.route('/services')
def services():
    return render_template('services.html')

@app.route('/contact')
def contact():
    return render_template('contact.html')

@app.route('/testimonials')
def testimonials():
    return render_template('testimonials.html')

@app.route('/articles')
def articles():
    return render_template('articles.html')

@app.route('/articles/article1')
def article1():
    return render_template('article1.html')

    

@app.route('/articles/article1/subarticle')
def subarticle():
    return render_template('subarticle.html')

@app.route('/passwords')
def passwords():
    return render_template('passwords.html')

@app.route('/submit_email', methods=['POST'])
def submit_email():
    email = request.form.get('email')
    message = request.form.get('message')
    # This line is vulnerable to XSS
    return f"Your message was: {message}<br>Your email was: {email}"



@app.route('/sqli', methods=['GET', 'POST'])
def sqli():
    if request.method == 'POST':
        username = request.form['username']
        # Vulnerable to SQL injection
        query = f"SELECT * FROM user WHERE username = '{username}'"
        result = db.engine.execute(query)
        user = result.fetchone()
        return f"User: {user}" if user else "No user found."

    return render_template('sqli.html')

if __name__ == '__main__':
    app.run(debug=True)
