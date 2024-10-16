(defcolumns X Y)

(vanish test1
        (- X (if Y 0)))

(vanish test2
        (- X (ifnot Y 16)))
