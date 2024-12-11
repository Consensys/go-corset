(defcolumns X (Y :@loob))

(defconstraint test ()
        (- X (if Y 0 16)))
