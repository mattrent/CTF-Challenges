from CTFd.models import Challenges, db
from CTFd.plugins import register_plugin_assets_directory
from CTFd.plugins.challenges import CHALLENGE_CLASSES, BaseChallenge
from CTFd.plugins.container_challenges.decay import DECAY_FUNCTIONS, logarithmic
from CTFd.plugins.migrations import upgrade
from flask import Blueprint, session
from CTFd.utils.decorators import authed_only
import requests
import urllib.parse
import os
import urllib3

urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)


class ContainerChallenge(Challenges):
    __mapper_args__ = {"polymorphic_identity": "container"}
    id = db.Column(
        db.Integer, db.ForeignKey("challenges.id", ondelete="CASCADE"), primary_key=True
    )
    initial = db.Column(db.Integer, default=0)
    minimum = db.Column(db.Integer, default=0)
    decay = db.Column(db.Integer, default=0)
    function = db.Column(db.String(32), default="logarithmic")

    def __init__(self, *args, **kwargs):
        super(ContainerChallenge, self).__init__(**kwargs)
        self.value = kwargs["initial"]


class ContainerValueChallenge(BaseChallenge):
    id = "container"  # Unique identifier used to register challenges
    name = "container"  # Name of a challenge type
    templates = (
        {  # Handlebars templates used for each aspect of challenge editing & viewing
            "create": "/plugins/container_challenges/assets/create.html",
            "update": "/plugins/container_challenges/assets/update.html",
            "view": "/plugins/container_challenges/assets/view.html",
        }
    )
    scripts = {  # Scripts that are loaded when a template is loaded
        "create": "/plugins/container_challenges/assets/create.js",
        "update": "/plugins/container_challenges/assets/update.js",
        "view": "/plugins/container_challenges/assets/view.js",
    }
    # Route at which files are accessible. This must be registered using register_plugin_assets_directory()
    route = "/plugins/container_challenges/assets/"
    # Blueprint used to access the static_folder directory.
    blueprint = Blueprint(
        "container_challenges",
        __name__,
        template_folder="templates",
        static_folder="assets",
    )
    challenge_model = ContainerChallenge

    @classmethod
    def calculate_value(cls, challenge):
        f = DECAY_FUNCTIONS.get(challenge.function, logarithmic)
        value = f(challenge)
        challenge.value = value
        db.session.commit()
        return challenge

    @classmethod
    def read(cls, challenge):
        """
        This method is in used to access the data of a challenge in a format processable by the front end.

        :param challenge:
        :return: Challenge object, data dictionary to be returned to the user
        """
        challenge = ContainerChallenge.query.filter_by(id=challenge.id).first()
        data = super().read(challenge)
        data.update(
            {
                "initial": challenge.initial,
                "decay": challenge.decay,
                "minimum": challenge.minimum,
                "function": challenge.function,
            }
        )
        return data

    @classmethod
    def update(cls, challenge, request):
        """
        This method is used to update the information associated with a challenge. This should be kept strictly to the
        Challenges table and any child tables.

        :param challenge:
        :param request:
        :return:
        """
        data = request.form or request.get_json()

        for attr, value in data.items():
            # We need to set these to floats so that the next operations don't operate on strings
            if attr in ("initial", "minimum", "decay"):
                value = float(value)
            setattr(challenge, attr, value)

        return ContainerValueChallenge.calculate_value(challenge)

    @classmethod
    def solve(cls, user, team, challenge, request):
        super().solve(user, team, challenge, request)

        ContainerValueChallenge.calculate_value(challenge)


def load(app):
    upgrade(plugin_name="container_challenges")
    CHALLENGE_CLASSES["container"] = ContainerValueChallenge
    app.db.create_all()
    register_plugin_assets_directory(
        app, base_path="/plugins/container_challenges/assets/"
    )

    backend_url = os.environ["BACKENDURL"]


    def get_token():
        token = session['token']
        return token['access_token']


    @app.route("/containers/<challenge_id>/status", methods=["GET"])
    @authed_only
    def challenge_status(challenge_id):
        token = get_token()
        headers = {"Authorization": f"Bearer {token}"}
        url = urllib.parse.urljoin(backend_url, "challenges/" + str(challenge_id) + "/status")
        print("request", url)
        response = requests.get(url, json={}, headers=headers, verify=False)
        print(response)
        return response.json(), response.status_code


    @app.route("/containers/<challenge_id>/start", methods=["POST"])
    @authed_only
    def challenge_start(challenge_id):
        token = get_token()
        headers = {"Authorization": f"Bearer {token}"}
        url = urllib.parse.urljoin(backend_url, "challenges/" + str(challenge_id) + "/start")
        print("request", url)
        response = requests.post(url, json={}, headers=headers, verify=False)
        print(response)
        return response.json(), response.status_code


    @app.route("/containers/<challenge_id>/stop", methods=["POST"])
    @authed_only
    def challenge_stop(challenge_id):
        token = get_token()
        headers = {"Authorization": f"Bearer {token}"}
        url = urllib.parse.urljoin(backend_url, "challenges/" + str(challenge_id) + "/stop")
        print("request", url)
        response = requests.post(url, json={}, headers=headers, verify=False)
        print(response)
        return response.json(), response.status_code
