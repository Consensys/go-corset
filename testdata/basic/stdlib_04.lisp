(defcolumns (X :binary@prove) (Y :i16))

(defconstraint old ()
  (eq! Y
       (+ (prev Y)
          (* X (- X (prev X))))))

(defconstraint new ()
  (if (or! (eq! X 0) (remained-constant! X))
           ;; == 0
           (remained-constant! Y)
           ;; == 1
           (did-inc! Y 1)))
