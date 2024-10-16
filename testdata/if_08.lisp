(defcolumns X Y Z)
(vanish test (if X (- Z (if Y 0))))
(vanish test (if X (- Z (ifnot Y 16))))
