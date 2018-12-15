module URI
  class Postgres < Generic
    DEFAULT_SCHEME = 'postgres'
    DEFAULT_PORT = 5432
    COMPONENT = %i[ scheme userinfo host port path query].freeze

    def self.build(args)
      tmp = Util.make_components_hash(self, args)
      super(tmp)
    end
  end

  @@schemes['POSTGRES'] = Postgres
end
