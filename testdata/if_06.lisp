(defcolumns X Y)

(defconstraint test1 ()
        (- X (if Y 0)))

(defconstraint test2 ()
        (- X (if Y 0 16)))
