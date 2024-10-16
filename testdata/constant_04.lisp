(defcolumns X Y)
;; X + 2 == Y + 2
(vanish c1 (- (+ X 2) (+ Y (^ 2 1))))
;; X + 4 == Y + 4
(vanish c1 (- (+ X 4) (+ Y (^ 2 2))))
;; X + 8 == Y + 8
(vanish c1 (- (+ X 8) (+ Y (^ 2 3))))
;; X + 16 == Y + 16
(vanish c1 (- (+ X 16) (+ Y (^ 2 4))))
;; X + 32 == Y + 32
(vanish c1 (- (+ X 32) (+ Y (^ 2 5))))
;; X + 64 == Y + 64
(vanish c1 (- (+ X 64) (+ Y (^ 2 6))))
;; X + 128 == Y + 128
(vanish c1 (- (+ X 128) (+ Y (^ 2 7))))
;; X + 256 == Y + 256
(vanish c1 (- (+ X 256) (+ Y (^ 2 8))))
