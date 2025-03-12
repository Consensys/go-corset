;;error:8:1-32:malformed declaration
;;error:9:17-21:expected identifier
;;error:10:1-22:malformed declaration
;;error:11:22-23:expected column declarations
(defcolumns (P :binary@prove) (A :i16))

(defperspective p1 P ((B :byte)))
(defperspective p2 ((B :byte)))
(defperspective (p3) P ((B :byte)))
(defperspective p4 P)
(defperspective p5 P A)
