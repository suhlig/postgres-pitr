require_relative '../uri/postgres'
require 'yaml'
require 'pathname'
require 'securerandom'

module PITR
  module Config
    class DB
      attr_reader :password

      def initialize(path)
        @config = YAML.load_file(path).fetch('db').merge('password' => read_password)
      end

      def user
        @config.fetch('user', nil)
      end

      def host
        @config.fetch('host', 'localhost')
      end

      def port
        @config.fetch('port', URI::Postgres::DEFAULT_PORT)
      end

      def name
        @config.fetch('name', nil)
      end

      def params
        @config.fetch('params', nil)
      end

      def url
        URI::Postgres.build(
          userinfo: [user, password].join(':'),
          host: host,
          port: port,
          path: '/' + name,
          query: params&.map{|kv| kv.join('=') }&.join('&'),
        )
      end

      private

      def read_password
        password_file = Pathname(__dir__) / '../../ansible/.postgres-password'

        if password_file.exist?
          password_file.read.chomp
        else
          SecureRandom.urlsafe_base64(32).tap do |password|
            password_file.write(password)
          end
        end
      end
    end
  end
end
