require_relative '../uri/postgres'
require 'yaml'
require 'pathname'
require 'securerandom'

module PITR
  module Config
    class Base
      attr_reader :config

      def initialize(path)
        @config = YAML.load_file(path)
      end
    end

    class DB < Base
      def user
        db.fetch('user')
      end

      def host
        db.fetch('host', 'localhost')
      end

      def port
        db.fetch('port', URI::Postgres::DEFAULT_PORT)
      end

      def name
        db.fetch('name')
      end

      def version
        db.fetch('version')
      end

      def password
        db.fetch('password')
      end

      def params
        db.fetch('params', {})
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

      def db
        config.fetch('db')
      end

    end

    class Minio < Base
      def port
        minio.fetch('port', 9000)
      end

      def access_key
        minio.fetch('access_key')
      end

      def secret_key
        minio.fetch('secret_key')
      end

      private

      def minio
        config.fetch('minio')
      end
    end
  end
end
