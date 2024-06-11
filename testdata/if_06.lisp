(column X)
(column Y)

(vanish test1
        (- Y (if X 3)))

(vanish test2
        (- Y (ifnot X 16)))
