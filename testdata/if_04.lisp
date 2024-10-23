(defcolumns X Y)

(defconstraint test ()
        (- X (if Y 0 16)))
