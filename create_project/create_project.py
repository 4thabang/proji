#!.env/bin/python3

import os
import sqlite3
import subprocess
import sys

from .helper import Helper


class CreateProject:

    # Path to the database
    _conf_dir = "/home/niko/.config/create_project/"
    _db = _conf_dir + "db/cp.sqlite"

    def __init__(self, project_name, lang):
        if type(project_name) != str:
            raise TypeError("Project name must be a string.")
        if type(lang) != str:
            raise TypeError("Language name must be a string.")
        self._project_name = project_name
        self._lang = lang
        self._lang_id = 0
        self._conn = None
        self._cur = None

    def run(self):
        # Check if directory already exists
        if self._does_dir_exist():
            return 1

        # Connect to database
        with sqlite3.connect(CreateProject._db) as self.conn:
            if not self.conn:
                err_msg = Helper.format_err_msg(
                    "Could not connect to database.")
                print(err_msg)
                return 2

            # Create cursor
            self._cur = self.conn.cursor()

            if not self._cur:
                return 3

            # Check if provided language is supported
            if not self._is_lang_supported():
                return 4

            # Create the project folder
            if not self._create_project_folder():
                return 5

            # Create language specific sub folders
            if not self._create_sub_folders():
                return 6

            # Create language specific files
            if not self._create_files():
                return 7

            # Copy template files
            if not self._copy_templates():
                return 8
        return 0

    def get_project_name(self):
        ''' Get the project name. '''
        return self._project_name

    def get_language(self):
        ''' Get the language. '''
        return self._lang

    @staticmethod
    def get_db_path():
        return CreateProject._db

    @staticmethod
    def get_config_dir_path():
        return CreateProject._conf_dir

    def _does_dir_exist(self):
        ''' Check if directory already exists. '''

        if os.path.exists(self._project_name):
            err_msg = Helper.format_err_msg("Directory already exists.")
            print(err_msg)
            return True
        return False

    def _is_lang_supported(self):
        ''' Check if the provided language is supported. '''

        for lang_short in self._cur.execute('''
                                            SELECT
                                                language_id,
                                                name_short
                                            FROM
                                                languages_short
                                            '''):
            if lang_short[1] == self._lang:
                # Found language
                self._lang_id = lang_short[0]
                return True

        # Language not supported
        langs = self._cur.execute(
            'SELECT name_short FROM languages_short').fetchall()

        err_msg = Helper.format_err_msg(
            "You have to specify a supported language.",
            ("Currently supported languages: " + str(langs)))
        print(err_msg)
        return False

    def _create_project_folder(self):
        ''' Create the main project directory. '''

        # Create main directory
        print("Creating project folder...")
        try:
            subprocess.run(["mkdir", "-p", self._project_name],
                           timeout=10.0, check=True)
        except (subprocess.CalledProcessError, TimeoutError) as err:
            print(err)
            return False
        return True

    def _create_sub_folders(self):
        ''' Create sub folders depending on the specified language. '''

        # Create subfolders
        print("Creating subfolders...")

        try:
            for sub_folder in self._cur.execute(
                '''
                SELECT
                    relative_dest_path
                FROM
                    folders
                WHERE
                    language_id=?
                ''',
                    (self._lang_id,)):

                sub_folder = self._project_name + "/" + sub_folder[0]
                subprocess.run(["mkdir",
                                "-p",
                                sub_folder],
                               timeout=10.0,
                               check=True)
        except (subprocess.CalledProcessError, TimeoutError) as err:
            print(err)
            return False
        return True

    def _create_files(self):
        ''' Create language specific files. '''
        # Create files
        print("Creating files...")

        try:
            for file in self._cur.execute(
                '''
                SELECT
                    relative_dest_path
                FROM
                    files
                WHERE
                    language_id=? and
                    is_template=?
                ''',
                    (self._lang_id, 0,)):

                file = self._project_name + "/" + file[0]
                subprocess.run(["touch", file], timeout=10.0, check=True)
        except (subprocess.CalledProcessError, TimeoutError) as err:
            print(err)
            return False
        return True

    def _copy_templates(self):
        ''' Create language specific files. '''
        # Create files
        print("Copying templates...")

        try:
            for template in self._cur.execute(
                '''
                SELECT
                    relative_dest_path,
                    absolute_orig_path
                FROM
                    files
                WHERE
                    language_id=? and
                    is_template=?''',
                    (self._lang_id, 1,)):

                dest = self._project_name + "/" + template[0]
                template = CreateProject._conf_dir + template[1]
                subprocess.run(["cp", template, dest],
                               timeout=30.0, check=True)
        except (subprocess.CalledProcessError, TimeoutError) as err:
            print(err)
            return False
        return True