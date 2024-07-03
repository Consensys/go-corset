(module m1)
(column X)

(module m2)
(column Y)
;;
(lookup test (Y) (m1.X))
