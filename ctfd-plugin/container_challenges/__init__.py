from CTFd.forms import BaseForm
from CTFd.models import Challenges, db
from CTFd.plugins import register_plugin_assets_directory
from CTFd.plugins.challenges import CHALLENGE_CLASSES, BaseChallenge
from CTFd.plugins.migrations import upgrade
from CTFd.utils.decorators import admins_only
from flask import Blueprint, request, render_template, session
from CTFd.utils.decorators import authed_only
import jwt
import requests
import urllib.parse
from wtforms import StringField, HiddenField
from wtforms.validators import InputRequired
import os
import datetime
import urllib3

urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)


class ContainerChallenge(Challenges):
    __mapper_args__ = {"polymorphic_identity": "container"}
    id = db.Column(
        db.Integer, db.ForeignKey("challenges.id", ondelete="CASCADE"), primary_key=True
    )
    identifier = db.Column(db.String(32), default="identifier")

    def __init__(self, *args, **kwargs):
        super(ContainerChallenge, self).__init__(**kwargs)


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
    challenge_model = ContainerChallenge


    @classmethod
    def read(cls, challenge):
        """
        This method is in used to access the data of a challenge in a format processable by the front end.

        :param challenge:
        :return: Challenge object, data dictionary to be returned to the user
        """
        challenge = ContainerChallenge.query.filter_by(id=challenge.id).first()
        data = {
            "id": challenge.id,
            "name": challenge.name,
            "value": challenge.value,
            "identifier": challenge.identifier,
            "description": challenge.description,
            "connection_info": challenge.connection_info,
            "next_id": challenge.next_id,
            "category": challenge.category,
            "state": challenge.state,
            "max_attempts": challenge.max_attempts,
            "type": challenge.type,
            "type_data": {
                "id": cls.id,
                "name": cls.name,
                "templates": cls.templates,
                "scripts": cls.scripts,
            },
        }
        return data


def load(app):
    upgrade(plugin_name="container_challenges")
    CHALLENGE_CLASSES["container"] = ContainerValueChallenge
    app.db.create_all()
    register_plugin_assets_directory(
        app, base_path="/plugins/container_challenges/assets/"
    )

    backend_url = os.getenv("BACKENDURL", "http://deployer.default.svc.cluster.local:8080/")
    jwt_secret = os.environ["jwtsecret"]

    def create_token():
        return jwt.encode({
            "userid": str(session["id"]),
            "role": "player",
            "exp": int((datetime.datetime.now() + datetime.timedelta(days=1)).timestamp())
            }, jwt_secret, algorithm="HS256")


    @app.route("/containers/<challenge_id>/status", methods=["GET"])
    @authed_only
    def challenge_status(challenge_id):
        token = create_token()
        headers = {"Authorization": f"Bearer {token}"}
        url = urllib.parse.urljoin(backend_url, "challenges/" + str(challenge_id) + "/status")
        print("request", url)
        response = requests.get(url, json={}, headers=headers, verify=False)
        print(response)
        return response.json()


    @app.route("/containers/<challenge_id>/start", methods=["POST"])
    # @authed_only
    def challenge_start(challenge_id):
        token = create_token()
        headers = {"Authorization": f"Bearer {token}"}
        url = urllib.parse.urljoin(backend_url, "challenges/" + str(challenge_id) + "/start")
        print("request", url)
        response = requests.post(url, json={}, headers=headers, verify=False)
        print(response)
        return response.json()


    @app.route("/containers/<challenge_id>/stop", methods=["POST"])
    @authed_only
    def challenge_stop(challenge_id):
        token = create_token()
        headers = {"Authorization": f"Bearer {token}"}
        url = urllib.parse.urljoin(backend_url, "challenges/" + str(challenge_id) + "/stop")
        print("request", url)
        response = requests.post(url, json={}, headers=headers, verify=False)
        print(response)
        return response.json()
